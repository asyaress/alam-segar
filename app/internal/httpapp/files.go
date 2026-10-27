package httpapp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type savedUpload struct {
	RelativePath string
	SHA256       string
	MimeType     string
}

func (a *App) savePrivateUpload(r *http.Request, field, directory string, maxBytes int64, required bool) (savedUpload, error) {
	file, header, err := r.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) && !required {
		return savedUpload{}, nil
	}
	if err != nil {
		return savedUpload{}, errors.New("file wajib dipilih")
	}
	defer file.Close()
	if header.Size <= 0 || header.Size > maxBytes {
		return savedUpload{}, fmt.Errorf("ukuran file harus antara 1 byte dan %d MB", maxBytes/(1024*1024))
	}
	return a.persistUpload(file, header, directory, maxBytes)
}

func (a *App) persistUpload(file multipart.File, header *multipart.FileHeader, directory string, maxBytes int64) (savedUpload, error) {
	return a.persistValidatedUpload(file, header, a.cfg.PrivateFilesRoot, directory, maxBytes, true)
}

func (a *App) persistPublicImage(file multipart.File, header *multipart.FileHeader, directory string, maxBytes int64) (savedUpload, error) {
	return a.persistValidatedUpload(file, header, a.cfg.PublicFilesRoot, directory, maxBytes, false)
}

func (a *App) persistValidatedUpload(file multipart.File, header *multipart.FileHeader, root, directory string, maxBytes int64, allowPDF bool) (savedUpload, error) {
	probe := make([]byte, 512)
	n, err := file.Read(probe)
	if err != nil && !errors.Is(err, io.EOF) {
		return savedUpload{}, err
	}
	probe = probe[:n]
	mimeType := http.DetectContentType(probe)
	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
	}
	if allowPDF {
		extensions["application/pdf"] = ".pdf"
	}
	ext, ok := extensions[mimeType]
	if !ok {
		return savedUpload{}, errors.New("format file harus JPG, PNG, atau PDF")
	}
	if strings.Contains(strings.ToLower(header.Filename), ".php") || strings.Contains(strings.ToLower(header.Filename), ".exe") {
		return savedUpload{}, errors.New("nama file tidak diizinkan")
	}
	random := make([]byte, 20)
	if _, err := rand.Read(random); err != nil {
		return savedUpload{}, err
	}
	relative := filepath.ToSlash(filepath.Join(directory, hex.EncodeToString(random)+ext))
	absolute := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(absolute), 0700); err != nil {
		return savedUpload{}, err
	}
	destination, err := os.OpenFile(absolute, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return savedUpload{}, err
	}
	defer destination.Close()
	hash := sha256.New()
	reader := io.MultiReader(strings.NewReader(string(probe)), io.LimitReader(file, maxBytes-int64(len(probe))+1))
	written, err := io.Copy(io.MultiWriter(destination, hash), reader)
	if err != nil {
		return savedUpload{}, err
	}
	if written > maxBytes {
		_ = destination.Close()
		_ = os.Remove(absolute)
		return savedUpload{}, errors.New("ukuran file melebihi batas")
	}
	return savedUpload{RelativePath: relative, SHA256: hex.EncodeToString(hash.Sum(nil)), MimeType: mimeType}, nil
}

func (a *App) servePrivateFile(w http.ResponseWriter, r *http.Request, relative string) {
	clean := filepath.Clean(filepath.FromSlash(relative))
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		http.NotFound(w, r)
		return
	}
	absolute := filepath.Join(a.cfg.PrivateFilesRoot, clean)
	root, _ := filepath.Abs(a.cfg.PrivateFilesRoot)
	resolved, _ := filepath.Abs(absolute)
	if !strings.HasPrefix(strings.ToLower(resolved), strings.ToLower(root+string(os.PathSeparator))) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Disposition", `inline; filename="dokumen"`)
	http.ServeFile(w, r, resolved)
}
