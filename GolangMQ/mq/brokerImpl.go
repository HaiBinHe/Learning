package mq

import (
	"errors"
	"sync"
	"time"
)

type BrokerImpl struct{
	exit chan bool	//关闭通道
	capacity int	//消息队列的容量
	//一个topic可以有多个订阅者
	topics map[string][]chan interface{} // key： topic  value ： queue
	sync.RWMutex
}

func NewBroker() *BrokerImpl{
	return &BrokerImpl{
		exit: make(chan bool),
		capacity: 100,
		topics: make(map[string][]chan interface{}),
	}
}

func (b *BrokerImpl) publish(topic string, msg interface{}) error {

	select {
	case <- b.exit:
		return errors.New("broker closed")
	default:

	}
	b.RLock()
	//是否有这个主题
	subscribe, ok := b.topics[topic]
	b.RUnlock()
	if !ok {
		return nil
	}
	//将msg广播到每个订阅者上
	b.broadcast(msg, subscribe)
	return nil
}
/*
为订阅的主题创建一个channel，然后将订阅者加入到对应的topic中就可以了，并且返回一个接收channel。
*/
func (b *BrokerImpl) subscribe(topic string) (<-chan interface{}, error) {
	select {
	case <- b.exit:
		return nil, errors.New("broker closed")
	default:
	}
	ch := make(chan interface{}, b.capacity)
	b.Lock()
	b.topics[topic] = append(b.topics[topic], ch)
	b.Unlock()
	return ch, nil
}
/*
取消订阅
sub: 要取消的订阅
*/
func (b *BrokerImpl) unsubscribe(topic string, sub <-chan interface{}) error {
	select {
	case <- b.exit:
		return errors.New("broker closed")
	default:
	}
	b.RLock()
	subs, ok := b.topics[topic]
	b.RUnlock()
	if !ok {
		return nil
	}
	var newsubs []chan interface{}
	for _, s := range subs{
		if s == sub{
			continue
		}
		newsubs = append(newsubs, s)
	}
	b.Lock()
	b.topics[topic] = newsubs
	b.Unlock()
	return nil
}
/*
关闭的同时 将消息队列清空
*/
func (b *BrokerImpl) close() {
	select {
	case <- b.exit:
		return
	default:
		close(b.exit)
		b.Lock()
		//这里主要是为了保证下一次使用该消息队列不发生冲突。
		b.topics = make(map[string][]chan interface{})
		b.Unlock()
	}
	return
}
/*
broadcast 一个topic有多个订阅者,根据当前订阅者的数量决定开启多少个goroutine
*/
func (b *BrokerImpl) broadcast(msg interface{}, subscribers []chan interface{}) {
	count := len(subscribers)
	concurrency := 1
	switch  {
	case count > 1000:
		concurrency = 3
	case count > 100:
		concurrency = 2
	default:
		concurrency = 1
	}
	pub := func(start int) {
		for j := start; j < count; j += concurrency{
			select {
			case subscribers[j] <- msg:
				//超过5毫秒就停止推送，接着进行下面的推送
			case <-time.After(time.Millisecond * 5):
			case <-b.exit:
				return
			}
		}
	}
	for i:=0; i < concurrency;i++{
		go pub(i)
	}
}

func (b *BrokerImpl) setConditions(capacity int) {
	b.capacity = capacity
}


