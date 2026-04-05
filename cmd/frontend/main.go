// 独立静态前端：在项目根目录执行
//
//	go run ./cmd/frontend -port 3000
//
// 将 public/config.js 里 API_ORIGIN 设为后端地址（如 http://127.0.0.1:8080），再访问 http://127.0.0.1:3000/
package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

func main() {
	port := flag.Int("port", 3000, "HTTP listen port")
	dir := flag.String("dir", "public", "directory containing index.html and assets")
	flag.Parse()

	abs, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir(abs))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			http.ServeFile(w, r, filepath.Join(abs, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	}))

	addr := ":" + strconv.Itoa(*port)
	log.Printf("frontend: serving %q at http://127.0.0.1%s (set public/config.js API_ORIGIN to your API)", abs, addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
