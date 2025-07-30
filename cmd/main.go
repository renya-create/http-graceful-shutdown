package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	err := run()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
func run() error {
	// 1. シグナル待ち受けのコンテキストを作成
	// SIGTERM, SIGINT, SIGKILLなどのシグナルを捕捉する
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	defer stop() // シグナルハンドラのリソースを解放

	// 2. HTTPサーバーのセットアップ
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // 長時間実行されるリクエストをシミュレートするために5秒間遅延
		fmt.Fprintln(w, "Hello, graceful shutdown!")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 3. http.Serveをgoroutineで起動
	go func() {
		log.Println("HTTP server started on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// 4. シグナルを受け取るまで待機
	<-ctx.Done()
	log.Println("Signal received. Starting graceful shutdown...")

	// 5. Graceful Shutdownの実行
	// 実行中の処理（コネクションなど）が完了するのを待つ
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 5秒のタイムアウトを設定
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error occurred during graceful shutdown: %v", err)
	}

	// 6. シャットダウン時に必要な追加処理（例: キャッシュの整理、データ保存など）
	log.Println("Executing additional shutdown tasks...")
	time.Sleep(2 * time.Second) // Example: 2 second processing
	log.Println("Additional tasks completed.")

	log.Println("HTTP server stopped successfully.")
	return nil
}
