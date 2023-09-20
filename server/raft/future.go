package raft

type Response struct {
	val any
	err error
}

type Future interface {
	Get() (any, error)
}

type MutableFuture interface {
	Future
	Set(val any, err error)
}

type BlockingFuture struct {
	resChan chan Response
}

func (f *BlockingFuture) Set(val any, err error) {
	f.resChan <- Response{val, err}
}

func (f *BlockingFuture) Get() (any, error) {
	res := <-f.resChan
	return res.val, res.err
}

func NewBlockingFuture() *BlockingFuture {
	return &BlockingFuture{
		resChan: make(chan Response, 1),
	}
}

type Operation[T any] struct {
	cmd   T
	reply MutableFuture
}

type RPC Operation[any]
type OfferRequest Operation[Command]
