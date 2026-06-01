package main

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"net"
	"strings"

	"github.com/migeru111/mkvs/kvs"
)

func main() {
	// 1. TCPサーバーをポート8080で起動
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		slog.Error("サーバーの起動に失敗しました: %v", err)
	}
	defer listener.Close()
	slog.Debug("サーバーがポート :8080 で待機中...")

	for {
		// fmt.Println("%i回目のループ", i)
		// 2. クライアントからの接続を待つ（接続されるまでここでブロック）
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("接続の受け入れに失敗しました: %v", err)
			continue
		}
		slog.Debug("クライアントが接続しました: %s\n", conn.RemoteAddr().String())

		// 3. 各クライアントの通信をgoroutine（分身）で並行処理
		handleClient(conn)
	}
}

// クライアントとの一連のやり取りを処理する関数
func handleClient(conn net.Conn) {
	defer conn.Close() // 関数が終わったら確実に接続を閉じる

	store := kvs.NewHashMapMutex()
	// クライアントからのデータを読み込むバッファを用意
	reader := bufio.NewReader(conn)

	for {
		// 改行コード（'\n'）までデータを読み込む
		message, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("クライアントが切断しました (%s): %v\n", conn.RemoteAddr().String(), err)
			return
		}

		// 届いたメッセージをトリミングして表示
		// メッセージ想定
		// GET: GET({key_name})
		// SET: SET({key_name},{value})
		// DELETE: DELETE({key_name})
		cleanMessage := strings.TrimSpace(message)
		ret := handleMessage(cleanMessage, store)
		//fmt.Printf("[%s] 受信: %s\n", conn.RemoteAddr().String(), cleanMessage)
		slog.Debug(ret)
		// オウム返し（Echo）のデータをクライアントに送信
		//response := fmt.Sprintf("サーバーからの返信: %s\n", cleanMessage)
		_, err = conn.Write([]byte(ret))
		if err != nil {
			log.Printf("データの送信に失敗しました: %v", err)
			return
		}
	}
}

func handleMessage(message string, store kvs.KVS) string {
	slog.Debug(message)
	if strings.HasPrefix(message, "GET") {
		key := strings.Trim(message, "GET()")
		value, _ := store.Get(key)
		return value + "\n"
	} else if strings.HasPrefix(message, "SET") {
		slog.Debug("message: %s", message)
		keyvalue := strings.Split(strings.Trim(message, "SET()"), ",")
		slog.Debug("keyvalue: %s", keyvalue)
		key, value := keyvalue[0], keyvalue[1]
		store.Set(key, value)
		return fmt.Sprintf("SET: %s,%s\n", key, value)
	} else {
		return fmt.Sprintf("unsupportedString: %s\n", message)
	}
}
