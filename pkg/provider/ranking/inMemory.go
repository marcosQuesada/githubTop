package ranking

import (
	"container/heap"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"sync"
)

const DefaultPriorityQueueSize = 10000

type InMemory struct {
	priorityQueue PriorityQueue
	maxSize       int
	index         map[string]*provider.Location
	mutex         sync.RWMutex
}

func NewInMemory(size int) *InMemory {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	return &InMemory{
		priorityQueue: pq,
		maxSize:       size,
		index:         make(map[string]*provider.Location),
	}
}

func (i *InMemory) IncreaseScore(city string) error {
	log.Infof("Increasing score from %s", city)
	i.mutex.Lock()
	defer i.mutex.Unlock()

	v, ok := i.index[city]
	if !ok {
		item := &provider.Location{
			Name:  city,
			Score: 1,
		}
		heap.Push(&i.priorityQueue, item)
		i.index[city] = item
		if i.priorityQueue.Len() > i.maxSize {
			_ = i.priorityQueue.Pop()
		}

		return nil
	}

	prio := v.Score + 1
	i.priorityQueue.update(v, v.Name, prio)

	return nil
}

func (i *InMemory) cap() {
	if i.priorityQueue.Len() <= i.maxSize {
		return
	}

	i.priorityQueue = i.priorityQueue[:i.maxSize]
}

func (i *InMemory) Top(size int) ([]*provider.Location, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if i.priorityQueue.Len() < size {
		size = i.priorityQueue.Len()
	}
	return i.priorityQueue[:size], nil
}

func (i *InMemory) Len() int {
	return i.priorityQueue.Len()
}

// A PriorityQueue implements heap.Interface and holds Locations.
type PriorityQueue []*provider.Location

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Score > pq[j].Score
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*provider.Location)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *provider.Location, value string, priority int) {
	item.Name = value
	item.Score = priority
	heap.Fix(pq, item.Index)
}
