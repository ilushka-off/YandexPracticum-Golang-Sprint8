package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	store := NewParcelStore(db)
	parcel := getTestParcel()

	assert.NoError(t, err)
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)

	assert.NoError(t, err)
	assert.NotEmpty(t, number)
	parcel.Number = number

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	data, err := store.Get(number)
	assert.NoError(t, err)

	assert.Equal(t, parcel.Client, data.Client)
	assert.Equal(t, parcel.Number, data.Number)
	assert.Equal(t, parcel.Address, data.Address)
	assert.Equal(t, parcel.Status, data.Status)
	assert.Equal(t, parcel.CreatedAt, data.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(number)
	assert.NoError(t, err)

	_, err = store.Get(number)
	assert.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	assert.NoError(t, err)
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)

	assert.NoError(t, err)
	assert.NotEmpty(t, number)
	parcel.Number = number

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)

	assert.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	data, err := store.Get(number)
	assert.NoError(t, err)

	assert.Equal(t, newAddress, data.Address)

}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	assert.NoError(t, err)
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)

	assert.NoError(t, err)
	assert.NotEmpty(t, number)
	parcel.Number = number

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(number, ParcelStatusRegistered)
	assert.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	data, err := store.Get(number)
	assert.NoError(t, err)

	assert.Equal(t, ParcelStatusRegistered, data.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")

	store := NewParcelStore(db)

	assert.NoError(t, err)
	defer db.Close()

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		assert.NoError(t, err)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	assert.NoError(t, err)
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	assert.Equal(t, len(parcels), len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		expected, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, expected, parcel)
	}
}
