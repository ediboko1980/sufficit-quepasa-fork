package models

type QPStore struct {
	BotID     string `db:"bot_id"`
	Data      []byte `db:"data"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type IQPStore interface {
	Create(wid string) (QPStore, error)
	Get(wid string) (QPStore, error)
	GetOrCreate(wid string) (QPStore, error)
	Update(wid string, data []byte) ([]byte, error)
	Delete(wid string) error
	Exists(wid string) (bool, error)
}
