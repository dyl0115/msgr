package cmd

import (
	"fmt"
	"io"
	"net"
	"syscall"
	"unsafe"

	"github.com/spf13/cobra"
)

var listenPort string

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "메시지 수신 대기 (팝업으로 알림)",
	Run: func(cmd *cobra.Command, args []string) {
		// 콘솔 창 숨기기
		hideConsole()

		addr := fmt.Sprintf(":%s", listenPort)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			showPopup("msgr 오류", fmt.Sprintf("리스닝 실패: %v", err))
			return
		}
		defer ln.Close()

		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			go handleMessage(conn)
		}
	},
}

func hideConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")

	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	showWindow := user32.NewProc("ShowWindow")

	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd != 0 {
		showWindow.Call(hwnd, 0) // SW_HIDE = 0
	}
}

func handleMessage(conn net.Conn) {
	defer conn.Close()

	buf, err := io.ReadAll(conn)
	if err != nil {
		return
	}

	message := string(buf)
	showPopup("📨 새 메시지", message)
}

func showPopup(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	msgPtr, _ := syscall.UTF16PtrFromString(message)

	messageBox.Call(
		0,
		uintptr(unsafe.Pointer(msgPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		0x00000040, // MB_ICONINFORMATION
	)
}

func init() {
	listenCmd.Flags().StringVar(&listenPort, "port", "9999", "리스닝 포트")
}
