package mq

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestClient(t *testing.T){
	c := NewClient()
	c.SetConditions(100)
	var wg sync.WaitGroup

	for i:=0;i < 100;i++{
		topic := fmt.Sprintf("GolangText(%d)", i)
		payload := fmt.Sprintf("hhb(%d)", i)

		ch, err := c.Subscribe(topic)
		if err != nil {
			t.Fatal(err)
		}
		wg.Add(1)
		go func() {
			e := c.GetPayLoad(ch)
			if e != payload{
				t.Fatalf("%s expectd %s but get %s",topic, payload, e)
			}
			if err := c.UnSubscribe(topic, ch); err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
		if err := c.Publish(topic, payload);err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
}
// 使用一个定时器，向一个主题定时推送消息.
var topic = "test"
func TestOnceTopic(t *testing.T)  {
	m := NewClient()
	m.SetConditions(10)
	ch,err :=m.Subscribe(topic)
	if err != nil{
		fmt.Println("subscribe failed")
		return
	}
	go OncePub(m)
	OnceSub(ch,m)
	defer m.Close()
}

// 定时推送
func OncePub(c *Client)  {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for  {
		select {
		case <- t.C:
			err := c.Publish(topic,"<-push Test->")
			if err != nil{
				fmt.Println("pub message failed")
			}
		default:

		}
	}
}

// 接受订阅消息
func OnceSub(m <-chan interface{},c *Client)  {
	for  {
		val := c.GetPayLoad(m)
		fmt.Printf("get message is %s\n",val)
	}
}

//多个topic测试
func TestClientTopic(t *testing.T)  {
	m := NewClient()
	defer m.Close()
	m.SetConditions(10)
	top := ""
	for i:=0;i<10;i++{
		//top 主题
		top = fmt.Sprintf("Golang_%02d",i)
		go Sub(m,top)
	}
	ManyPub(m)
}

func ManyPub(c *Client)  {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for  {
		select {
		case <- t.C:
			for i:= 0;i<10;i++{
				//多个topic 推送不同的消息
				top := fmt.Sprintf("Golang_%02d",i)
				payload := fmt.Sprintf("asong_%02d",i)
				err := c.Publish(top,payload)
				if err != nil{
					fmt.Println("pub message failed")
				}
			}
		default:

		}
	}
}

func Sub(c *Client,top string)  {
	ch,err := c.Subscribe(top)
	if err != nil{
		fmt.Printf("sub top:%s failed\n",top)
	}
	for  {
		val := c.GetPayLoad(ch)
		if val != nil{
			fmt.Printf("%s get message is %s\n",top,val)
		}
	}
}

