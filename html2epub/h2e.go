package html2epub

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gonejack/get"

	"github.com/gonejack/html-to-epub/go-epub"
)

type HtmlToEpub struct {
	Options
	DefaultCover []byte

	book   *epub.Epub
	imgIdx int
}

func (h *HtmlToEpub) Run() (err error) {
	if h.About {
		fmt.Println("Visit https://github.com/gonejack/html-to-epub")
		return
	}
	_, exx := os.Stat(h.Output)
	if exx == nil {
		return fmt.Errorf("output file %s already exist", h.Output)
	}
	if len(h.HTML) == 0 {
		return errors.New("no .html file given")
	}
	return h.run()
}
func (h *HtmlToEpub) run() (err error) {
	err = h.makeBook()
	if err != nil {
		return
	}

	refs := make(map[string]string)
	for i, html := range h.HTML {
		err = h.add(i+1, refs, html)
		if err != nil {
			err = fmt.Errorf("parse %s failed: %s", html, err)
			return
		}
	}

	err = h.book.Write(h.Output)
	if err != nil {
		return fmt.Errorf("cannot write output epub: %s", err)
	}

	return
}
func (h *HtmlToEpub) makeBook() error {
	h.book = epub.NewEpub(h.Title)
	h.book.SetAuthor(h.Author)
	h.book.SetDescription(fmt.Sprintf("Epub generated at %s with github.com/gonejack/html-to-epub", time.Now().Format("2006-01-02")))
	if h.Css!="" {
		h.book.AddCSS(h.Css, "custom_style.css")
	}
	return h.setCover()
}
func (h *HtmlToEpub) setCover() (err error) {
	if h.Cover == "" {
		temp, err := os.CreateTemp("", "html-to-epub")
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

	m, err := mimetype.DetectFile(h.Cover)
	if err != nil {
		return fmt.Errorf("cannot detect cover mime type %s", err)
	}
	coverRef, err := h.book.AddImage(h.Cover, "cover"+m.Extension())
	if err != nil {
		return fmt.Errorf("cannot add cover %s", err)
	}
	h.book.SetCover(coverRef, "")

	return
}
func (h *HtmlToEpub) add(index int, refs map[string]string, html string) (err error) {
	fd, err := os.Open(html)
	if err != nil {
		return
	}
	defer fd.Close()

	doc, err := goquery.NewDocumentFromReader(fd)
	if err != nil {
		return
	}
	doc = h.cleanDoc(doc)

	images := h.saveImages(doc)
	doc.Find("img").Each(func(i int, img *goquery.Selection) { h.changeRef(html, img, refs, images) })

	title := doc.Find("title").Text()
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(html), filepath.Ext(html))
	}
	title = fmt.Sprintf("%d. %s", index, title)

	content, err := doc.Find("body").Html()
	if err != nil {
		return
	}
	
	cssInternalPath := ""
	if h.Css != "" {
		cssInternalPath = "../css/custom_style.css"
	}
	
	_, err = h.book.AddSection(content, title, "", cssInternalPath)

	return
}
func (h *HtmlToEpub) saveImages(doc *goquery.Document) map[string]string {
	downloads := make(map[string]string)

	tasks := get.NewDownloadTasks()
	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		src, _ := img.Attr("src")
		if !strings.HasPrefix(src, "http") {
			return
		}

		localFile, exist := downloads[src]
		if exist {
			return
		}

		uri, err := url.Parse(src)
		if err != nil {
			log.Printf("parse %s fail: %s", src, err)
			return
		}
		_ = os.MkdirAll(h.ImagesDir, 0766)
		localFile = filepath.Join(h.ImagesDir, fmt.Sprintf("%s%s", md5str(src), filepath.Ext(uri.Path)))

		tasks.Add(src, localFile)
		downloads[src] = localFile
	})
	get.Batch(tasks, 3, time.Minute*2).ForEach(func(t *get.DownloadTask) {
		if t.Err != nil {
			log.Printf("download %s fail: %s", t.Link, t.Err)
		}
	})

	return downloads
}
func (h *HtmlToEpub) changeRef(htmlFile string, img *goquery.Selection, refs, downloads map[string]string) {
	img.RemoveAttr("loading")
	img.RemoveAttr("srcset")

	src, _ := img.Attr("src")

	internalRef, exist := refs[src]
	if exist {
		img.SetAttr("src", internalRef)
		return
	}

	var localFile string
	switch {
	case strings.HasPrefix(src, "data:"):
		return
	case strings.HasPrefix(src, "http"):
		localFile, exist = downloads[src]
		if !exist {
			log.Printf("local file of %s not exist", src)
			return
		}
	default:
		fd, err := h.openLocalFile(htmlFile, src)
		if err != nil {
			log.Printf("local ref %s not found: %s", src, err)
			return
		}
		_ = fd.Close()
		localFile = fd.Name()
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
	internalName := fmt.Sprintf("image_%03d", h.imgIdx)
	{
		h.imgIdx += 1
		if !strings.HasSuffix(internalName, fmime.Extension()) {
			internalName += fmime.Extension()
		}
		internalRef, err = h.book.AddImage(localFile, internalName)
		if err != nil {
			log.Printf("cannot add image %s: %s", localFile, err)
			return
		}
		refs[src] = internalRef
	}

	if h.Verbose {
		log.Printf("replace %s as %s", src, localFile)
	}

	img.SetAttr("src", internalRef)
}
func (h *HtmlToEpub) openLocalFile(htmlFile string, ref string) (fd *os.File, err error) {
	fd, err = os.Open(ref)
	if err == nil {
		return
	}

	// compatible with evernote's exported htmls
	dirname := strings.TrimSuffix(htmlFile, filepath.Ext(htmlFile))
	name := filepath.Base(ref)
	fd, err = os.Open(filepath.Join(dirname+"_files", name))
	if err == nil {
		return
	}
	fd, err = os.Open(filepath.Join(dirname+".resources", name))
	if err == nil {
		return
	}
	if strings.HasSuffix(ref, ".") {
		return h.openLocalFile(htmlFile, strings.TrimSuffix(ref, "."))
	}

	return
}
func (h *HtmlToEpub) cleanDoc(doc *goquery.Document) *goquery.Document {
	// remove inoreader ads
	doc.Find("body").Find(`div:contains("ads from inoreader")`).Closest("center").Remove()

	// remove solidot.org ads
	doc.Find("img[src='https://img.solidot.org//0/446/liiLIZF8Uh6yM.jpg']").Remove()

	return doc
}

func md5str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
