package server

import (
	"bytes"
	"github.com/klauspost/compress/zip"
	"github.com/klauspost/compress/zlib"
	"github.com/valyala/fasthttp"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	errorMessageOpeningFile       = "Error opening file"
	errorMessageCreatingFile      = "Error creating file"
	errorMessageCreatingDirectory = "Error creating directory"
	dirUserFiles                  = "./user_files/"
)

func handleFileRoute(ctx *fasthttp.RequestCtx) {
	if ctx.IsGet() {
		handleGetRequest(ctx)
	} else if ctx.IsPost() {
		handlePostRequest(ctx)
	}
}

func handleErrors(message string, err error, ctx *fasthttp.RequestCtx) bool {
	if err != nil {
		log.Printf("%s: %v", message, err)
		ctx.Error(message, fasthttp.StatusInternalServerError)
		return true
	}
	return false
}

func handleGetRequest(requestCtx *fasthttp.RequestCtx) {
	if requestCtx.QueryArgs().Has("q") {
		filename := requestCtx.QueryArgs().Peek("q")
		f := filepath.Join("data", string(filename)) + ".zip"
		file, err := os.Open(f)
		if handleErrors(errorMessageOpeningFile, err, requestCtx) {
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)

		// Set the necessary headers
		requestCtx.Response.Header.Set("Content-Disposition", "attachment; filename="+f)
		requestCtx.Response.Header.Set("Content-Type", "application/octet-stream")
		_, err = io.Copy(requestCtx.Response.BodyWriter(), file)
		if handleErrors(errorMessageOpeningFile, err, requestCtx) {
			return
		}
	}
}

func handlePostRequest(requestCtx *fasthttp.RequestCtx) {
	// open directory
	err := os.MkdirAll(dirUserFiles, os.ModePerm)
	if handleErrors(errorMessageCreatingDirectory, err, requestCtx) {
		return
	}

	// extract the filename from the header
	header := string(requestCtx.Request.Header.Peek("Content-Disposition"))
	parts := strings.Split(header, ";")
	filename := strings.Trim(parts[len(parts)-1], " ")
	filename = filepath.Join(dirUserFiles, filename)

	file, err := os.Create(filename)
	if handleErrors(errorMessageCreatingFile, err, requestCtx) {
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = io.Copy(file, bytes.NewReader(requestCtx.PostBody()))
	if handleErrors(errorMessageCreatingFile, err, requestCtx) {
		return
	}

	var buf bytes.Buffer
	zipW := zip.NewWriter(&buf)

	// set the best compression
	zipW.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return zlib.NewWriterLevel(out, zlib.BestCompression)
	})

	// attach the file to zip
	f, err := zipW.Create(path.Base(file.Name()))
	if err != nil {
		log.Printf("Error creating zip entry: %v", err)
		requestCtx.Error("Error creating zip entry", fasthttp.StatusInternalServerError)
		err := zipW.Close()
		if err != nil {
			return
		}
		return
	}

	// fill zip content with file data
	data, err := os.ReadFile(file.Name())
	if err != nil {
		log.Printf("Error reading file for the zip: %v", err)
		requestCtx.Error("Error reading file for the zip", fasthttp.StatusInternalServerError)
		err := zipW.Close()
		if err != nil {
			return
		}
		return
	}

	_, err = f.Write(data)
	if err != nil {
		log.Printf("Error writing file to the zip: %v", err)
		requestCtx.Error("Error writing file to the zip", fasthttp.StatusInternalServerError)
		err := zipW.Close()
		if err != nil {
			return
		}
		return
	}

	// Add metadata to the manifest
	manifest, err := os.OpenFile("./user_files/manifest.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Printf("Error opening manifest file: %v", err)
		requestCtx.Error("Error opening manifest file", fasthttp.StatusInternalServerError)
		return
	}
	defer func(manifest *os.File) {
		_ = manifest.Close()
	}(manifest)

	_, err = manifest.WriteString("File name: " + file.Name() + " Size: " + strconv.Itoa(len(data)) + "\n")
	if err != nil {
		log.Printf("Error writing to manifest file: %v", err)
		requestCtx.Error("Error writing to manifest file", fasthttp.StatusInternalServerError)
		return
	}

	zipW.Close()
}
