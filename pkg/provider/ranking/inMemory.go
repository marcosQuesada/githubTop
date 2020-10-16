package ranking

import (
	"container/heap"
	"context"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"github.com/marcosQuesada/githubTop/pkg/provider"
	"sync"
)

// DefaultPriorityQueueSize max priority queue size
const DefaultPriorityQueueSize = 10000

// InMemory defines an inMemory ranking in top of a priority queue over the heap
type InMemory struct {
	priorityQueue PriorityQueue
	maxSize       int
	index         map[string]*provider.Location
	mutex         sync.RWMutex
}

// NewInMemory instantiates inMemory ranking
func NewInMemory(size int) *InMemory {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	return &InMemory{
		priorityQueue: pq,
		maxSize:       size,
		index:         make(map[string]*provider.Location),
	}
}

// IncreaseScore city score  increase by 1
func (i *InMemory) IncreaseScore(_ context.Context, city string) error {
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

// Top returns priority queue from head up to "size" length
func (i *InMemory) Top(_ context.Context, size int) ([]*provider.Location, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if i.priorityQueue.Len() < size {
		size = i.priorityQueue.Len()
	}
	return i.priorityQueue[:size], nil
}

// Len returns ranking size
func (i *InMemory) Len(_ context.Context) (int64, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	s := i.priorityQueue.Len()

	return int64(s), nil
}

// A PriorityQueue implements heap.Interface and holds Locations.
type PriorityQueue []*provider.Location

// Len returns priority queue size
func (pq PriorityQueue) Len() int { return len(pq) }

// Less comparative method between locations
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Score > pq[j].Score
}

// Swap exchanges 2 locations
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

// Push inserts a location in the top of the priority queue
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*provider.Location)
	item.Index = n
	*pq = append(*pq, item)
}

// Pop remove and returns last element on priority queue
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) update(item *provider.Location, value string, priority int) {
	item.Name = value
	item.Score = priority
	heap.Fix(pq, item.Index)
}
