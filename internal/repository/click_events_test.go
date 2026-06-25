package repository

import (
	"context"
	"fmt"
	"shorter-url/internal/domain"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func Test_Create_CreateEvent_Pass(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	ctx := context.Background()
	dateExample := time.Now()

	inputData := &domain.ClickEvent{
		ShortUrlId: 1,
		IpAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		Referer:    "https://google.com",
	}

	mockRows := pgxmock.NewRows([]string{"id", "ip_address", "short_url_id", "user_agent", "referer", "clicked_at"}).
		AddRow(int64(42), inputData.IpAddress, inputData.ShortUrlId, inputData.UserAgent, inputData.Referer, dateExample)

	queryRegex := `^INSERT INTO click_events\s*\(short_url_id, ip_address, user_agent, referer\)\s*VALUES\s*\(\$1, \$2, \$3, \$4\)\s*RETURNING id, ip_address, short_url_id, user_agent, referer, clicked_at$`

	mockPool.ExpectQuery(queryRegex).
		WithArgs(inputData.ShortUrlId, inputData.IpAddress, inputData.UserAgent, inputData.Referer).
		WillReturnRows(mockRows)

	repo := NewClickEventsRepository(mockPool)
	result, err := repo.Create(ctx, inputData)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(42), result.Id)
	assert.Equal(t, dateExample, result.ClickedAt)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func Test_Create_CreateEvent_Fail(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	ctx := context.Background()
	inputData := &domain.ClickEvent{
		ShortUrlId: 1,
		IpAddress:  "192.168.1.1",
	}

	queryRegex := `^INSERT INTO click_events.*`
	mockPool.ExpectQuery(queryRegex).
		WithArgs(inputData.ShortUrlId, inputData.IpAddress, inputData.UserAgent, inputData.Referer).
		WillReturnError(fmt.Errorf("database connection closed"))

	repo := NewClickEventsRepository(mockPool)
	result, err := repo.Create(ctx, inputData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "something wrong when create data in click event")

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func Test_FindByShortUrlId_Pass(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	ctx := context.Background()
	var targetShortUrlId int64 = 99
	var targetUserId int64 = 1 // Tambahkan mock User ID yang merequest rabs
	dateExample := time.Now()

	mockRows := pgxmock.NewRows([]string{"id", "short_url_id", "ip_address", "user_agent", "referer", "clicked_at"}).
		AddRow(int64(1), targetShortUrlId, "192.168.1.1", "Chrome", "Direct", dateExample).
		AddRow(int64(2), targetShortUrlId, "192.168.1.2", "Safari", "https://github.com", dateExample)

	queryRegex := `(?i)^SELECT\s+ce\.id,\s+ce\.short_url_id,\s+ce\.ip_address,\s+ce\.user_agent,\s+ce\.referer,\s+ce\.clicked_at\s+FROM\s+click_event\s+ce\s+INNER\s+JOIN\s+short_urls\s+su\s+ON\s+ce\.short_url_id\s+=\s+su\.id\s+WHERE\s+ce\.short_url_id\s+=\s+\$1\s+AND\s+su\.user_id\s+=\s+\$2$`

	mockPool.ExpectQuery(queryRegex).
		WithArgs(targetShortUrlId, targetUserId). // Wajib menyertakan targetUserId ($2)
		WillReturnRows(mockRows)

	repo := NewClickEventsRepository(mockPool)
	// Masukkan targetUserId ke dalam parameter panggil ketiga
	results, err := repo.FindByShortUrlId(ctx, targetShortUrlId, targetUserId)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, targetShortUrlId, results[0].ShortUrlId)
	assert.Equal(t, "Chrome", results[0].UserAgent)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func Test_FindByShortUrlId_Fail_QueryError(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	ctx := context.Background()
	var targetShortUrlId int64 = 99
	var targetUserId int64 = 1

	// Regex fleksibel menangkap query SELECT yang gagal
	queryRegex := `(?i)^SELECT\s+ce\.id,\s+ce\.short_url_id.*`

	mockPool.ExpectQuery(queryRegex).
		WithArgs(targetShortUrlId, targetUserId). // Samakan argumennya rabs
		WillReturnError(fmt.Errorf("syntax error"))

	repo := NewClickEventsRepository(mockPool)
	results, err := repo.FindByShortUrlId(ctx, targetShortUrlId, targetUserId)

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "something wrong when find click events by short code")

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
