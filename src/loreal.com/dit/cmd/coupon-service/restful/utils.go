package restful

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"net/http"
)

//outputJSON - output json for http response
func outputJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	enc := json.NewEncoder(w)
	if DEBUG {
		enc.SetIndent("", " ")
	}
	if err := enc.Encode(data); err != nil {
		log.Println("[ERR] - JSON encode error:", err)
		http.Error(w, "500", http.StatusInternalServerError)
		return
	}
}

//outputGzipJSON - output json for http response
func outputGzipJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip")
	// zw, _ := gzip.NewWriterLevel(w, gzip.BestCompression)
	zw := gzip.NewWriter(w)
	defer func(zw *gzip.Writer) {
		zw.Flush()
		zw.Close()
	}(zw)
	enc := json.NewEncoder(zw)
	if DEBUG {
		enc.SetIndent("", " ")
	}
	if err := enc.Encode(data); err != nil {
		log.Println("[ERR] - JSON encode error:", err)
		http.Error(w, "500", http.StatusInternalServerError)
		return
	}
}
