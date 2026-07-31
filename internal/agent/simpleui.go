package agent

import "bytes"
import "context"
import "errors"
import "fmt"
import "io"

import "myagent/internal/handler"
import "myagent/internal/session"
import "myagent/internal/session/userinput"
import "myagent/internal/session/response"
import "myagent/pkg/logger"


func StartSimpleUI(
	input io.Reader,
	output io.Writer,
	sessionID string,
) error {
	msgQueue := userinput.NewMessageQueue()
	och := make(chan *response.AgentResponse, 65536) // emit 事件容量
	go readMessage(input, msgQueue)
	go outputMessage(output, och)


	a, err := NewAgent(sessionID, msgQueue, och)
	if err != nil {
		fmt.Printf("NewAgent failed: %v\n", err)
		return err
	}

	for {
		logger.Debugf("test\n")
		fmt.Fprint(output, "User>")

		ctx, cancel := context.WithCancel(context.Background())
		clientState := &session.ClientState{CancelFunc: cancel}


		userInput := msgQueue.GetInput()

		// check client command
		if userInput.Type() == userinput.InputTypeCommand {
			res, err := handler.HandleClientCommand(
				ctx,
				clientState,
				string(userInput.Content),
			)
			if err == nil {
				logger.Debugf("exec client command: %v %v", res, err)
				fmt.Fprintf(output, "|🔧: %s\n", res)
				continue
			} else if !errors.Is(err, handler.SkipHandleCommand) {
				fmt.Fprintf(output, ">>>❗Error: %s\n", err.Error())
			}
		}

		go Exec(a, userInput)
	}

}

// readMessage 读取消息 r -> ch
func readMessage(r io.Reader, queue *userinput.MessageQueue) {
	delimiter := []byte("\n")
	// fmt.Printf("| delimiter: (%v)\n", delimiter)

	var data []byte
	var buf [1]byte
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			data = append(data, buf[0])

			if bytes.HasSuffix(data, delimiter) {
				data = data[:len(data)-len(delimiter)]

				queue.Push(&userinput.UserInput{Content: data})

				data = data[:0]
			}
		}

		if err != nil {
			logger.Errorf("read input error: %v\n", err)
		}
	}
}

// outputMessage 打印信息 ch -> w
func outputMessage(w io.Writer, ch chan *response.AgentResponse) {
	for content := range ch {
		switch content.RespType {
		case response.AgentRespTypeLLM:
			fmt.Fprintf(w, "Agent✨> %s\n", content.LLMResponse.Content())
			fmt.Fprintf(w, "<finished>\n")
		case response.AgentRespTypeError:
			fmt.Fprintf(w, ">>>❗Error: %s\n", content.Err.Error())
		case response.AgentRespTypeMiddleMsg:
			fmt.Fprintf(w, "|🤔> %s\n", content.MiddleMessage)
		case response.AgentRespTypeCmd:
			fmt.Fprintf(w, "|☁️🔧: %s\n", content.CmdResult)
		}
	}
}
