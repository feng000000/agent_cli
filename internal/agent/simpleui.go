package agent

import "bytes"
import "context"
import "errors"
import "fmt"
import "io"

import "myagent/internal/handler"
import "myagent/internal/runtime"
import "myagent/pkg/logger"

func StartSimpleUI(
	input io.Reader,
	output io.Writer,
	sessionID string,
) error {
	msgQueue := runtime.NewMessageQueue()
	och := make(chan *runtime.AgentResponse, 65536) // emit 事件容量
	go readMessage(input, msgQueue)
	go outputMessage(output, och)


	a, err := NewAgent(msgQueue, och, sessionID)
	if err != nil {
		return err
	}

	for {
		logger.Debugf("test\n")
		fmt.Fprint(output, "User>")

		ctx, cancel := context.WithCancel(context.Background())
		clientState := &runtime.ClientState{CancelFunc: cancel}


		userInput := msgQueue.GetInput()

		// check client command
		if userInput.Type() == runtime.InputTypeCommand {
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
func readMessage(r io.Reader, queue *runtime.MessageQueue) {
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

				queue.Push(&runtime.UserInput{Content: data})

				data = data[:0]
			}
		}

		if err != nil {
			logger.Errorf("read input error: %v\n", err)
		}
	}
}

// outputMessage 打印信息 ch -> w
func outputMessage(w io.Writer, ch chan *runtime.AgentResponse) {
	for content := range ch {
		switch content.RespType {
		case runtime.AgentRespTypeLLM:
			fmt.Fprintf(w, "Agent✨> %s\n", content.LLMResponse.Content())
		case runtime.AgentRespTypeError:
			fmt.Fprintf(w, ">>>❗Error: %s\n", content.Err.Error())
		case runtime.AgentRespTypeMiddleMsg:
			fmt.Fprintf(w, "|🤔> %s\n", content.MiddleMessage)
		case runtime.AgentRespTypeCmd:
			fmt.Fprintf(w, "|☁️🔧: %s\n", content.CmdResult)
		}
	}
}
