package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/solve", solveHandler)

	log.Printf("OCR Translate Service starting on port %s", port)
	log.Printf("Tesseract version: %s", getTesseractVersion())
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","engine":"tesseract","version":"%s"}`, getTesseractVersion())
}

func solveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing 'image' field: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read image: "+err.Error(), http.StatusBadRequest)
		return
	}

	psm := r.FormValue("psm")
	if psm == "" {
		psm = "7"
	}
	whitelist := r.FormValue("whitelist")
	if whitelist == "" {
		whitelist = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	result, err := solveCaptcha(imageData, psm, whitelist)
	if err != nil {
		log.Printf("solve error: %v", err)
		http.Error(w, "solve failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("solved: %q", result)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, result)
}

func solveCaptcha(imageData []byte, psm, whitelist string) (string, error) {
	tmpIn, err := os.CreateTemp("", "ocr-in-*.png")
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	defer os.Remove(tmpIn.Name())

	if _, err := tmpIn.Write(imageData); err != nil {
		tmpIn.Close()
		return "", fmt.Errorf("write temp: %w", err)
	}
	tmpIn.Close()

	tmpOut, err := os.CreateTemp("", "ocr-out-*")
	if err != nil {
		return "", fmt.Errorf("create temp out: %w", err)
	}
	outBase := tmpOut.Name()
	tmpOut.Close()
	os.Remove(outBase)
	defer os.Remove(outBase + ".txt")

	cmd := exec.Command("tesseract",
		tmpIn.Name(),
		outBase,
		"--psm", psm,
		"-c", "tessedit_char_whitelist="+whitelist,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract: %w, stderr: %s", err, stderr.String())
	}

	resultBytes, err := os.ReadFile(outBase + ".txt")
	if err != nil {
		return "", fmt.Errorf("read output: %w", err)
	}

	result := strings.TrimSpace(string(resultBytes))
	result = strings.ReplaceAll(result, " ", "")
	result = strings.ReplaceAll(result, "\n", "")

	if result == "" {
		return "", fmt.Errorf("empty result")
	}

	return result, nil
}

func getTesseractVersion() string {
	out, err := exec.Command("tesseract", "--version").Output()
	if err != nil {
		return "not found"
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return "unknown"
}
