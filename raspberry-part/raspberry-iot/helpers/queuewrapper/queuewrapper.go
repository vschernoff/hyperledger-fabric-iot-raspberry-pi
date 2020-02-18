package queuewrapper

import (
	"container/list"
	"fmt"
	"hlf-iot/config"
	"hlf-iot/helpers/ca"
	"hlf-iot/helpers/httpwrapper"
	"time"
)

type QueueWrapper struct {
	Queue *list.List `json:"queue"`
}

type SendData struct {
	Fcn  string   `json:"fcn"`
	Args []string `json:"args"`
}

type QueueStructure struct {
	GetDataFcn   func() (*SendData, error) `json:"getdatafcn"`
	PreparedData *SendData                 `json:"prepareddata"`
}

func Init() *QueueWrapper {
	queueWrapper := &QueueWrapper{}
	queueWrapper.Queue = list.New()

	return queueWrapper
}

func (queueWrapper *QueueWrapper) AddToQueue(queueElement *QueueStructure) {
	queueWrapper.Queue.PushBack(queueElement)
	fmt.Printf("Queue add\n")
	fmt.Printf("Queue length +1: %d\n", queueWrapper.Queue.Len())
	fmt.Println("========================================")
}

func (queueWrapper *QueueWrapper) StartDaemon() {
	for {
		element := queueWrapper.Queue.Front()
		if element != nil {
			var err error
			queue := element.Value.(*QueueStructure)
			var sendData *SendData
			if queue.GetDataFcn != nil {
				sendData, err = queue.GetDataFcn()
				if err != nil {
					panic(err.Error())
				}
			}
			if queue.PreparedData != nil {
				sendData = queue.PreparedData
			}
			fmt.Println("========================================")
			queueWrapper.Queue.Remove(element)
			if sendData != nil {
				fmt.Println()
				fmt.Printf("Sending data: %s", sendData)
				fmt.Println()
				err = queueWrapper.PrepareRequest(sendData, config.GPRS_API_URL)
				if err != nil {
					panic(err.Error())
				}
				fmt.Printf("Queue remove: %s\n", sendData)
			} else {
				fmt.Println("Remove without sending")
			}
			fmt.Printf("Queue length -1: %d\n", queueWrapper.Queue.Len())
			fmt.Println("========================================")
		}
		time.Sleep(config.DELAY_FOR_DAEMON_MILLISECONDS * time.Millisecond)
	}
}

func (queueWrapper *QueueWrapper) PrepareRequest(data *SendData, url string) error {
	fabricCa := ca.GetInstance()
	success := false

	for {
		proposalReq, err := fabricCa.ProposalReq(data.Fcn, data.Args)
		if err != nil {
			return err
		}

		success, err = httpwrapper.PostReq(proposalReq)
		if err != nil {
			return err
		}
		if !success {
			continue
		}

		err = fabricCa.SignProposal()
		if err != nil {
			return err
		}

		broadcastPayloadReq, err := fabricCa.BroadcastPayloadReq()
		if err != nil {
			return err
		}

		success, err = httpwrapper.PostReq(broadcastPayloadReq)
		if err != nil {
			return err
		}
		if !success {
			continue
		}

		err = fabricCa.SignBroadcastPayload()
		if err != nil {
			return err
		}

		broadcastReq, err := fabricCa.BroadcastReq()
		if err != nil {
			return err
		}

		success, err = httpwrapper.PostReq(broadcastReq)
		if err != nil {
			return err
		}
		if !success {
			continue
		} else {
			break
		}

		time.Sleep(2000 * time.Millisecond)
	}

	return nil
}
