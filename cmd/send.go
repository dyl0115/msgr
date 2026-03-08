package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
)

var toIP string
var port string

var sendCmd = &cobra.Command{
	Use:   "send [메시지]",
	Short: "다른 PC로 메시지 전송",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		message := args[0]
		addr := fmt.Sprintf("%s:%s", toIP, port)

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Printf("연결 실패: %v\n", err)
			return
		}
		defer conn.Close()

		_, err = fmt.Fprintf(conn, message)
		if err != nil {
			fmt.Printf("전송 실패: %v\n", err)
			return
		}

		fmt.Printf("메시지 전송 완료 → %s\n", addr)
	},
}

func init() {
	sendCmd.Flags().StringVar(&toIP, "to", "", "대상 PC IP (필수)")
	sendCmd.Flags().StringVar(&port, "port", "9999", "포트 번호")
	sendCmd.MarkFlagRequired("to")
}
