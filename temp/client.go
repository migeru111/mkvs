package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	// 1. サーバー（localhost:8080）に接続
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("サーバーへの接続に失敗しました: %v", err)
	}
	defer conn.Close()
	fmt.Println("サーバーに接続しました。メッセージを入力してください（'quit' で終了）:")

	// キーボード入力用のリーダー
	inputReader := bufio.NewReader(os.Stdin)
	// サーバーからの受信用のリーダー
	serverReader := bufio.NewReader(conn)

	for {
		fmt.Print("> ")
		// 2. キーボードからの入力を読み込む
		input, err := inputReader.ReadString('\n')
		if err != nil {
			log.Fatalf("入力の読み込みに失敗しました: %v", err)
		}

		// 'quit' が入力されたら終了
		if strings.TrimSpace(input) == "quit" {
			fmt.Println("接続を終了します。")
			break
		}

		// 3. サーバーへデータを送信（末尾に改行が必要）
		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Fatalf("サーバーへの送信に失敗しました: %v", err)
		}

		// 4. サーバーからの返信を受信して表示
		response, err := serverReader.ReadString('\n')
		if err != nil {
			log.Fatalf("サーバーからの受信に失敗しました: %v", err)
		}
		fmt.Print(response)
	}
}
