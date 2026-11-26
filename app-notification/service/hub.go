package service

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	pushChannel = "notify:push"
	ticketTTL   = 60 * time.Second
)

// PushEvent 是推给浏览器的一条实时事件。
type PushEvent struct {
	PrincipalType int    `json:"principalType"`
	PrincipalID   uint   `json:"principalId"`
	Event         string `json:"event"`
	Payload       any    `json:"payload,omitempty"`
}

// Principal 是票据还原出的主体。
type Principal struct {
	Type int  `json:"type"`
	ID   uint `json:"id"`
}

// Hub 把实时事件推到本机连接，并用 Redis 扇出到其它实例。
type Hub struct {
	redis  redis.UniversalClient
	mu     sync.RWMutex
	conns  map[string]map[uint64]chan []byte
	nextID uint64
	once   sync.Once
	cancel context.CancelFunc
}

func NewHub(client redis.UniversalClient) *Hub {
	return &Hub{
		redis: client,
		conns: map[string]map[uint64]chan []byte{},
	}
}

// Start 订阅 Redis。重复调用只生效一次。
func (h *Hub) Start() {
	h.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		h.cancel = cancel
		go h.subscribe(ctx)
	})
}

// Stop 结束订阅。
func (h *Hub) Stop() {
	if h.cancel != nil {
		h.cancel()
	}
}

func (h *Hub) subscribe(ctx context.Context) {
	sub := h.redis.Subscribe(ctx, pushChannel)
	defer sub.Close()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var ev PushEvent
			if json.Unmarshal([]byte(msg.Payload), &ev) != nil {
				continue
			}
			h.deliverLocal(ev)
		}
	}
}

// Publish 发布一条实时事件。Redis 不可用时退回本机连接。
func (h *Hub) Publish(ctx context.Context, ev PushEvent) {
	raw, err := json.Marshal(ev)
	if err != nil {
		return
	}
	if err := h.redis.Publish(ctx, pushChannel, raw).Err(); err != nil {
		h.deliverLocal(ev)
	}
}

func (h *Hub) deliverLocal(ev PushEvent) {
	raw, err := json.Marshal(ev)
	if err != nil {
		return
	}
	if ev.PrincipalID == 0 {
		for _, group := range h.conns {
			for _, ch := range group {
				select {
				case ch <- raw:
				default:
				}
			}
		}
		return
	}
	key := connKey(ev.PrincipalType, ev.PrincipalID)
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.conns[key] {
		select {
		case ch <- raw:
		default:
		}
	}
}

// Subscribe 登记一条本机连接。返回的函数用于断开。
func (h *Hub) Subscribe(principalType int, principalID uint) (<-chan []byte, func()) {
	ch := make(chan []byte, 16)

	h.mu.Lock()
	h.nextID++
	id := h.nextID
	key := connKey(principalType, principalID)
	if h.conns[key] == nil {
		h.conns[key] = map[uint64]chan []byte{}
	}
	h.conns[key][id] = ch
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		delete(h.conns[key], id)
		if len(h.conns[key]) == 0 {
			delete(h.conns, key)
		}
		h.mu.Unlock()
	}
}

func connKey(principalType int, principalID uint) string {
	return strconv.Itoa(principalType) + ":" + strconv.FormatUint(uint64(principalID), 10)
}

// IssueTicket 签发一次性实时连接票据。
func (h *Hub) IssueTicket(ctx context.Context, p Principal) (string, error) {
	token := uuid.NewString()
	raw, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	if err := h.redis.Set(ctx, "notify:ticket:"+token, raw, ticketTTL).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// ConsumeTicket 校验并销毁票据。
func (h *Hub) ConsumeTicket(ctx context.Context, token string) (Principal, error) {
	raw, err := h.redis.GetDel(ctx, "notify:ticket:"+token).Result()
	if err != nil {
		return Principal{}, TicketInvalid()
	}
	var p Principal
	if json.Unmarshal([]byte(raw), &p) != nil || p.ID == 0 {
		return Principal{}, TicketInvalid()
	}
	return p, nil
}
