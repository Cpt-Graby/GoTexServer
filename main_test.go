package main

import (
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TEST GENERER A L'IA

func setTestEnv(t *testing.T) *os.Root {
	t.Helper()

	tmp := t.TempDir()
	files := filepath.Join(tmp, "files")

	mustWrite := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
			t.Fatalf("écriture de %s : %v", name, err)
		}
	}

	if err := os.MkdirAll(filepath.Join(files, "docs"), 0o755); err != nil {
		t.Fatalf("création des dossiers : %v", err)
	}
	mustWrite(filepath.Join(tmp, "secret.txt"), "TOP SECRET")
	mustWrite(filepath.Join(files, "hello.txt"), "Bonjour le monde")
	mustWrite(filepath.Join(files, "résumé.txt"), "cv")
	mustWrite(filepath.Join(files, "docs", "rapport.txt"), "rapport annuel")

	if err := os.Symlink(filepath.Join(tmp, "secret.txt"), filepath.Join(files, "evil")); err != nil {
		t.Fatalf("création du lien symbolique : %v", err)
	}

	root, err := os.OpenRoot(files)
	if err != nil {
		t.Fatalf("OpenRoot : %v", err)
	}
	t.Cleanup(func() { root.Close() })
	return root
}

func TestDownload(t *testing.T) {
	root := setTestEnv(t)
	mux := createMux(root) // le même routeur que dans main

	tests := []struct {
		name         string
		method       string
		target       string
		header       map[string]string
		wantStatus   int
		wantRedirect bool
		wantBody     string
		wantFile     string
	}{
		// --- Cas nominaux ---
		{name: "fichier existant", method: "GET", target: "/dl/hello.txt",
			wantStatus: http.StatusOK, wantBody: "Bonjour le monde", wantFile: "hello.txt"},
		{name: "fichier dans un sous-dossier", method: "GET", target: "/dl/docs/rapport.txt",
			wantStatus: http.StatusOK, wantBody: "rapport annuel", wantFile: "rapport.txt"},
		{name: "nom accentué", method: "GET", target: "/dl/r%C3%A9sum%C3%A9.txt",
			wantStatus: http.StatusOK, wantBody: "cv", wantFile: "résumé.txt"},
		{name: "requête Range", method: "GET", target: "/dl/hello.txt",
			header:     map[string]string{"Range": "bytes=0-6"},
			wantStatus: http.StatusPartialContent, wantBody: "Bonjour"},
		{name: "HEAD accepté", method: "HEAD", target: "/dl/hello.txt",
			wantStatus: http.StatusOK},

		// --- Ce qui doit être refusé ---
		{name: "fichier absent", method: "GET", target: "/dl/nope.txt",
			wantStatus: http.StatusNotFound},
		{name: "racine /dl/ (pas de listing)", method: "GET", target: "/dl/",
			wantStatus: http.StatusNotFound},
		{name: "sous-dossier sans fichier", method: "GET", target: "/dl/docs",
			wantStatus: http.StatusNotFound},
		{name: "sous-dossier avec / final", method: "GET", target: "/dl/docs/",
			wantStatus: http.StatusNotFound},
		{name: "lien symbolique sortant", method: "GET", target: "/dl/evil",
			wantStatus: http.StatusNotFound},
		{name: "POST refusé", method: "POST", target: "/dl/hello.txt",
			wantStatus: http.StatusMethodNotAllowed},

		// --- Path traversal ---
		{name: "traversal ..", method: "GET", target: "/dl/../secret.txt",
			wantRedirect: true},
		{name: "traversal .. encodé", method: "GET", target: "/dl/%2e%2e/secret.txt",
			wantStatus: http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, nil)
			for k, v := range tc.header {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			res := rec.Result()
			body, _ := io.ReadAll(res.Body)

			if tc.wantRedirect {
				if res.StatusCode < 300 || res.StatusCode > 399 {
					t.Fatalf("statut = %d, attendu une redirection (3xx)", res.StatusCode)
				}
			} else if res.StatusCode != tc.wantStatus {
				t.Fatalf("statut = %d, attendu %d", res.StatusCode, tc.wantStatus)
			}

			if string(body) == "TOP SECRET" {
				t.Fatal("le fichier secret a fuité !")
			}

			if tc.wantBody != "" && string(body) != tc.wantBody {
				t.Errorf("corps = %q, attendu %q", body, tc.wantBody)
			}

			if tc.wantFile != "" {
				cd := res.Header.Get("Content-Disposition")
				_, params, err := mime.ParseMediaType(cd)
				if err != nil {
					t.Fatalf("Content-Disposition illisible %q : %v", cd, err)
				}
				if params["filename"] != tc.wantFile {
					t.Errorf("filename = %q, attendu %q (en-tête : %q)", params["filename"], tc.wantFile, cd)
				}
				if got := res.Header.Get("X-Content-Type-Options"); got != "nosniff" {
					t.Errorf("X-Content-Type-Options = %q, attendu \"nosniff\"", got)
				}
			}
		})
	}
}

func TestPOSTAllowHeader(t *testing.T) {
	mux := createMux(setTestEnv(t))

	req := httptest.NewRequest(http.MethodPost, "/dl/hello.txt", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Header().Get("Allow"), "GET, HEAD"; got != want {
		t.Errorf("Allow = %q, attendu %q", got, want)
	}
}
