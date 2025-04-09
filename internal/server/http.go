package server

import (
    "log"
    "net/http"
    "telegramassist/internal/api"
)

func StartHTTPServer(alertHandler *api.AlertHandler) {
    http.HandleFunc("/api/alerts", alertHandler.HandleAlert)

    go func() {
 	   log.Println("Iniciando servidor HTTP en :8080...")
 	   if err := http.ListenAndServe(":8080", nil); err != nil {
 		   log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
 	   }
    }()
}