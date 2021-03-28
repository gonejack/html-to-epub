package cmd

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"

	"io"
	"log"
	"net/http"

	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bmaupin/go-epub"
	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type HtmlToEpub struct {
	client http.Client

	ImagesDir string

	DefaultCover []byte

	Cover   string
	Title   string
	Author  string
	Verbose bool

	book *epub.Epub
}

func (h *HtmlToEpub) Run(htmls []string, output string) (err error) {
	if len(htmls) == 0 {
		return errors.New("no html given")
	}

	err = h.mkdirs()
	if err != nil {
		return
	}

	h.book = epub.NewEpub(h.Title)
	{
		h.setAuthor()
		h.setDesc()
		err = h.setCover()
		if err != nil {
			return
		}
	}

	for i, html := range htmls {
		err = h.addHTML(i+1, html)
		if err != nil {
			err = fmt.Errorf("parse %s failed: %s", html, err)
			return
		}
	}

	err = h.book.Write(output)
	if err != nil {
		return fmt.Errorf("cannot write output epub: %s", err)
	}

	return
}

func (h *HtmlToEpub) setAuthor() {
	h.book.SetAuthor(h.Author)
}
func (h *HtmlToEpub) setDesc() {
	h.book.SetDescription(fmt.Sprintf("Epub generated at %s with github.com/gonejack/textbundle-to-epub", time.Now().Format("2006-01-02")))
}
func (h *HtmlToEpub) setCover() (err error) {
	if h.Cover == "" {
		temp, err := os.CreateTemp("", "textbundle-to-epub")
		if err != nil {
			return fmt.Errorf("cannot create tempfile: %s", err)
		}
		_, err = temp.Write(h.DefaultCover)
		if err != nil {
			return fmt.Errorf("cannot write tempfile: %s", err)
		}
		_ = temp.Close()

		h.Cover = temp.Name()
	}

	fmime, err := mimetype.DetectFile(h.Cover)
	if err != nil {
		return fmt.Errorf("cannot detect cover mime type %s", err)
	}
	coverRef, err := h.book.AddImage(h.Cover, "epub-cover"+fmime.Extension())
	if err != nil {
		return fmt.Errorf("cannot add cover %s", err)
	}
	h.book.SetCover(coverRef, "")

	return
}

func (h *HtmlToEpub) addHTML(index int, html string) (err error) {
	file, err := os.Open(html)
	if err != nil {
		return
	}

	document, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return
	}

	document = h.cleanDoc(document)
	downloads := h.downloadImages(document)

	savedRefs := make(map[string]string)
	document.Find("img").Each(func(i int, img *goquery.Selection) {
		h.changeRef(img, savedRefs, downloads)
	})

	title := document.Find("title").Text()
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(html), filepath.Ext(html))
	}
	title = fmt.Sprintf("%d. %s", index, title)

	content, err := document.Find("body").Html()
	if err != nil {
		return
	}

	_, err = h.book.AddSection(content, title, "", "")

	return
}
func (h *HtmlToEpub) downloadImages(doc *goquery.Document) map[string]string {
	downloadPaths := make(map[string]string)
	downloadLinks := make([]string, 0)
	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		src, _ := img.Attr("src")
		if !strings.HasPrefix(src, "http") {
			return
		}

		localFile, exist := downloadPaths[src]
		if exist {
			return
		}

		uri, err := url.Parse(src)
		if err != nil {
			log.Printf("parse %s fail: %s", src, err)
			return
		}
		localFile = filepath.Join(h.ImagesDir, fmt.Sprintf("%s%s", md5str(src), filepath.Ext(uri.Path)))

		downloadPaths[src] = localFile
		downloadLinks = append(downloadLinks, src)
	})

	var batch = semaphore.NewWeighted(3)
	var group errgroup.Group

	for i := range downloadLinks {
		_ = batch.Acquire(context.TODO(), 1)

		link := downloadLinks[i]
		group.Go(func() error {
			defer batch.Release(1)

			if h.Verbose {
				log.Printf("fetch %s", link)
			}

			err := h.download(link, downloadPaths[link])
			if err != nil {
				log.Printf("download %s fail: %s", link, err)
			}

			return nil
		})
	}

	_ = group.Wait()

	return downloadPaths
}
func (h *HtmlToEpub) download(src string, path string) (err error) {
	timeout, cancel := context.WithTimeout(context.TODO(), time.Minute*2)
	defer cancel()

	info, err := os.Stat(path)
	if err == nil {
		headReq, headErr := http.NewRequestWithContext(timeout, http.MethodHead, src, nil)
		if headErr != nil {
			return headErr
		}
		resp, headErr := h.client.Do(headReq)
		if headErr == nil && info.Size() == resp.ContentLength {
			return // skip download
		}
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	request, err := http.NewRequestWithContext(timeout, http.MethodGet, src, nil)
	if err != nil {
		return
	}
	response, err := h.client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	var written int64
	if h.Verbose {
		bar := progressbar.NewOptions64(response.ContentLength,
			progressbar.OptionSetTheme(progressbar.Theme{Saucer: "=", SaucerPadding: ".", BarStart: "|", BarEnd: "|"}),
			progressbar.OptionSetWidth(10),
			progressbar.OptionSpinnerType(11),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetDescription(filepath.Base(src)),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionClearOnFinish(),
		)
		defer bar.Clear()
		written, err = io.Copy(io.MultiWriter(file, bar), response.Body)
	} else {
		written, err = io.Copy(file, response.Body)
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("response status code %d invalid", response.StatusCode)
	}

	if err == nil && written < response.ContentLength {
		err = fmt.Errorf("expected %s but downloaded %s", humanize.Bytes(uint64(response.ContentLength)), humanize.Bytes(uint64(written)))
	}

	return
}
func (h *HtmlToEpub) changeRef(img *goquery.Selection, savedRefs, downloads map[string]string) {
	img.RemoveAttr("loading")
	img.RemoveAttr("srcset")

	src, _ := img.Attr("src")

	if strings.HasPrefix(src, "data:") {
		return
	}

	internalRef, exist := savedRefs[src]
	if exist {
		img.SetAttr("src", internalRef)
		return
	}

	var localFile string
	switch {
	case strings.HasPrefix(src, "http"):
		localFile, exist = downloads[src]
		if !exist {
			log.Printf("local file of %s not exist", src)
			return
		}
	default:
		localFile = src
		_, err := os.Stat(localFile)
		if err != nil {
			localFile, _ = url.QueryUnescape(localFile)
		}
	}

	if h.Verbose {
		log.Printf("replace %s as %s", src, localFile)
	}

	// check mime
	fmime, err := mimetype.DetectFile(localFile)
	{
		if err != nil {
			log.Printf("cannot detect image mime of %s: %s", src, err)
			return
		}
		if !strings.HasPrefix(fmime.String(), "image") {
			log.Printf("mime of %s is %s instead of images", src, fmime.String())
			return
		}
	}

	// add image
	internalName := filepath.Base(localFile)
	{
		if !strings.HasSuffix(internalName, fmime.Extension()) {
			internalName += fmime.Extension()
		}
		internalRef, err = h.book.AddImage(localFile, internalName)
		if err != nil {
			log.Printf("cannot add image %s", err)
			return
		}
		savedRefs[src] = internalRef
	}

	img.SetAttr("src", internalRef)
}
func (h *HtmlToEpub) cleanDoc(doc *goquery.Document) *goquery.Document {
	// remove inoreader ads
	doc.Find("body").Find(`div:contains("ads from inoreader")`).Closest("center").Remove()

	return doc
}
func (h *HtmlToEpub) mkdirs() error {
	err := os.MkdirAll(h.ImagesDir, 0777)
	if err != nil {
		return fmt.Errorf("cannot make images dir %s", err)
	}
	return nil
}
func md5str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
