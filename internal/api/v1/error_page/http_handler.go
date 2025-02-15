package error_page

import "net/http"

func GetHandleFunc() http.HandlerFunc {
	setup()
	return serveHTTP
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path == "/" {
		http.Error(w, "invalid path", http.StatusNotFound)
		return
	}
	content, ok := fileContentMap.Load(r.URL.Path)
	if !ok {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	w.Write(content)
}
