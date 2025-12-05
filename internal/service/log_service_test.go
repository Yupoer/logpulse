package service

import (
	"context"
	"testing"

	"github.com/Yupoer/logpulse/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---
// 我們不需要真的連 Redis/Kafka，只要模擬它們的行為

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendLog(ctx context.Context, entry *domain.LogEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	return nil
}

type MockCacheRepo struct {
	mock.Mock
}

func (m *MockCacheRepo) IncrementLogCount(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *MockCacheRepo) GetLogCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return int64(args.Int(0)), args.Error(1)
}
func (m *MockCacheRepo) SetLog(ctx context.Context, entry *domain.LogEntry) error { return nil }
func (m *MockCacheRepo) GetLog(ctx context.Context, id uint) (*domain.LogEntry, error) {
	return nil, nil
}

type MockLogRepo struct{ mock.Mock }

func (m *MockLogRepo) Create(ctx context.Context, entry *domain.LogEntry) error { return nil }
func (m *MockLogRepo) GetByID(ctx context.Context, id uint) (*domain.LogEntry, error) {
	return nil, nil
}

type MockESRepo struct{ mock.Mock }

func (m *MockESRepo) BulkIndex(ctx context.Context, entries []*domain.LogEntry) error { return nil }
func (m *MockESRepo) Search(ctx context.Context, query string) ([]*domain.LogEntry, error) {
	return nil, nil
}

// --- Tests ---

func TestCreateLog(t *testing.T) {
	// 1. Setup
	mockProducer := new(MockProducer)
	mockCache := new(MockCacheRepo)
	mockLogRepo := new(MockLogRepo) // 不需要用到，但要傳進去
	mockESRepo := new(MockESRepo)   // 不需要用到，但要傳進去

	// 定義預期行為: 當 SendLog 被呼叫時，回傳 nil (成功)
	mockProducer.On("SendLog", mock.Anything, mock.Anything).Return(nil)
	// 定義預期行為: 當 Increment 被呼叫時，回傳 nil
	mockCache.On("IncrementLogCount", mock.Anything).Return(nil)
	// 定義預期行為: 當 GetLogCount 被呼叫時，回傳 100
	mockCache.On("GetLogCount", mock.Anything).Return(100, nil)

	service := NewLogService(mockProducer, mockLogRepo, mockCache, mockESRepo)

	// 2. Execute
	entry := &domain.LogEntry{ServiceName: "test", Message: "hello"}
	count, err := service.CreateLog(context.Background(), entry)

	// 3. Assert (驗證結果)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), count)

	// 驗證 Mock 物件是否真的被呼叫了
	mockProducer.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
