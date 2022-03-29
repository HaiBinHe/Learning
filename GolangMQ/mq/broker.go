package mq

type Broker interface {
	//publish 进行消息的推送，有两个参数即topic、msg，分别是订阅的主题、要传递的消息
	publish(topic string, msg interface{}) error
	//subscribe 消息的订阅，传入订阅的主题，即可完成订阅，并返回对应的channel通道用来接收数据
	subscribe(topic string)(<-chan interface{}, error)
	//unsubscribe 取消订阅
	unsubscribe(topic string, sub <-chan interface{}) error
	//关闭消息队列
	close()
	//broadcast 这个属于内部方法，作用是进行广播，对推送的消息进行广播，保证每一个订阅者都可以收到
	broadcast(msg interface{}, subscribers []chan interface{})
	//设置消息队列的容量
	setConditions(capacity int)
}
