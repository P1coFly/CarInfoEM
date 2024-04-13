package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/P1coFly/CarInfoEM/internal/models/car"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

// Функция для инициализации storage
func New(urlPath, portDB, userDB, password, nameDB, migrationsPath string) (*Storage, error) {
	const op = "storage.postgresql.New"

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		urlPath, portDB, userDB, password, nameDB)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверка соединения с базой данных
	if err := db.Ping(); err != nil {
		db.Close() // Закрыть соединение, если проверка не удалась
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		fmt.Println("no  migration file")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return &Storage{db: db}, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// Метод для регистрации авто
func (s *Storage) AddCar(car car.Car) (int, error) {
	const op = "storage.postgresql.AddCar"

	PeopleID, err := s.addPeople(car.Owner)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	var carID int
	err = s.db.QueryRow(`INSERT INTO CARS (reg_num, mark,model,year,owner_id) VALUES ($1, $2, $3, $4, $5) returning id`,
		car.RegNum, car.Mark, car.Model, car.Year, PeopleID).Scan(&carID)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}
	return carID, nil

}

// Метод для регистрации человека
func (s *Storage) addPeople(people car.People) (int, error) {
	const op = "storage.postgresql.AddPeople"

	var id int
	err := s.db.QueryRow(`INSERT INTO PEOPLES (name, surname, patronymic) VALUES ($1, $2, $3) returning id`,
		people.Name, people.Surname, people.Patronymic).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil

}

func (s *Storage) DeleteCar(carID int) (int, error) {
	const op = "storage.postgresql.DeleteCar"

	ownerID, err := s.getOwnerIDByCarID(carID)
	if err != nil {
		return ownerID, fmt.Errorf("%s: %w", op, err)
	}

	result, err := s.db.Exec(`DELETE FROM CARS WHERE id = $1`, carID)
	if err != nil {
		return carID, fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на количество удаленных записей
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return carID, fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return -1, fmt.Errorf("%s: %w", op, errors.New("car not found"))
	}

	_, err = s.db.Exec(`DELETE FROM PEOPLES WHERE id = $1`, ownerID)
	if err != nil {
		return ownerID, fmt.Errorf("%s: %w", op, err)
	}

	return carID, nil
}

func (s *Storage) getOwnerIDByCarID(carID int) (int, error) {
	const op = "storage.postgresql.getOwnerIDByCarID"
	row := s.db.QueryRow("SELECT owner_id FROM CARS WHERE id = $1", carID)

	var id int
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, fmt.Errorf("%s: %w", op, errors.New("car not found"))
		}
		return carID, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) GetCars(pageSize, pageToken int, carFilter car.CarFilter) ([]car.CarWithOwner, error) {
	const op = "storage.postgresql.GetCars"

	// Создаем базовый SQL-запрос
	sqlQuery := "SELECT CARS.ID, CARS.reg_num, CARS.mark, CARS.model, CARS.year, PEOPLES.name, PEOPLES.surname, PEOPLES.patronymic FROM CARS " +
		"JOIN PEOPLES ON CARS.owner_id = PEOPLES.id"

	// Формируем условия фильтрации, если они указаны
	var conditions []string

	if carFilter.YearFilter != "" {
		years := strings.Split(carFilter.YearFilter, ":")
		startYear := years[0]
		endYear := years[1]
		conditions = append(conditions, fmt.Sprintf("CARS.year BETWEEN %s AND %s", startYear, endYear))
	}

	if carFilter.RegNumFilter != "" {
		conditions = append(conditions, fmt.Sprintf("CARS.reg_num LIKE '%%%s%%'", carFilter.RegNumFilter))
	}

	if carFilter.ModelFilter != "" {
		conditions = append(conditions, fmt.Sprintf("CARS.model LIKE '%%%s%%'", carFilter.ModelFilter))
	}

	if carFilter.MarkFilter != "" {
		conditions = append(conditions, fmt.Sprintf("CARS.mark LIKE '%%%s%%'", carFilter.MarkFilter))
	}

	if carFilter.NameFilter != "" {
		conditions = append(conditions, fmt.Sprintf("PEOPLES.name LIKE '%%%s%%'", carFilter.NameFilter))
	}
	if carFilter.SurnameFilter != "" {
		conditions = append(conditions, fmt.Sprintf("PEOPLES.surname LIKE '%%%s%%'", carFilter.SurnameFilter))
	}
	if carFilter.PatronymicFilter != "" {
		conditions = append(conditions, fmt.Sprintf("PEOPLES.patronymic LIKE '%%%s%%'", carFilter.PatronymicFilter))
	}

	if len(conditions) > 0 {
		sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	sqlQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, (pageToken-1)*pageSize)

	rows, err := s.db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var cars []car.CarWithOwner
	for rows.Next() {
		cwo := car.CarWithOwner{}
		err := rows.Scan(&cwo.Id, &cwo.RegNum, &cwo.Mark, &cwo.Model, &cwo.Year, &cwo.Name, &cwo.Surname, &cwo.Patronymic)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		cars = append(cars, cwo)
	}

	return cars, nil
}

func (s *Storage) GetTotalCarsCount(carFilter car.CarFilter) (int, error) {
	const op = "storage.postgresql.GetTotalCarsCount"

	// Запрос для получения общего количества записей
	sqlQuery := "SELECT COUNT(DISTINCT CARS.ID) FROM CARS LEFT JOIN PEOPLES ON CARS.owner_id = PEOPLES.id"

	// Формируем условия фильтрации, если они указаны
	var conditions []string

	if carFilter.YearFilter != "" {
		years := strings.Split(carFilter.YearFilter, ":")
		startYear := years[0]
		endYear := years[1]
		conditions = append(conditions, fmt.Sprintf("CARS.year BETWEEN %s AND %s", startYear, endYear))
	}

	if carFilter.RegNumFilter != "" {
		conditions = append(conditions, fmt.Sprintf("CARS.reg_num LIKE '%%%s%%'", carFilter.RegNumFilter))
	}

	if carFilter.ModelFilter != "" {
		conditions = append(conditions, fmt.Sprintf("CARS.model LIKE '%%%s%%'", carFilter.ModelFilter))
	}

	if carFilter.MarkFilter != "" {
		conditions = append(conditions, fmt.Sprintf("CARS.mark LIKE '%%%s%%'", carFilter.MarkFilter))
	}

	if carFilter.NameFilter != "" {
		conditions = append(conditions, fmt.Sprintf("PEOPLES.name LIKE '%%%s%%'", carFilter.NameFilter))
	}
	if carFilter.SurnameFilter != "" {
		conditions = append(conditions, fmt.Sprintf("PEOPLES.surname LIKE '%%%s%%'", carFilter.SurnameFilter))
	}
	if carFilter.PatronymicFilter != "" {
		conditions = append(conditions, fmt.Sprintf("PEOPLES.patronymic LIKE '%%%s%%'", carFilter.PatronymicFilter))
	}

	if len(conditions) > 0 {
		sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	err := s.db.QueryRow(sqlQuery).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return totalCount, nil
}

func (s *Storage) PatchCar(carID int, pc car.PatchCar) (int, error) {
	const op = "storage.postgresql.PatchCar"

	code, err := s.patchOwner(carID, pc.PatchPeople)
	if err != nil {
		return code, err
	}

	carQuery := "UPDATE CARS SET "

	var params []interface{}
	var sql []string
	if pc.RegNum.Valid {
		sql = append(sql, fmt.Sprintf("reg_num = $%d,", len(sql)+1))
		params = append(params, pc.RegNum.String)
	}
	if pc.Mark.Valid {
		sql = append(sql, fmt.Sprintf("mark = $%d,", len(sql)+1))
		params = append(params, pc.Mark.String)
	}
	if pc.Model.Valid {
		sql = append(sql, fmt.Sprintf("model = $%d,", len(sql)+1))
		params = append(params, pc.Model.String)
	}
	if pc.Year.Valid {
		sql = append(sql, fmt.Sprintf("year = $%d,", len(sql)+1))
		params = append(params, pc.Year.Int16)
	}

	if len(sql) == 0 {
		return code, nil
	}
	carQuery += strings.Join(sql, " ")

	carQuery = carQuery[:len(carQuery)-1] + fmt.Sprintf(" WHERE id = $%d", len(sql)+1)
	params = append(params, carID)

	result, err := s.db.Exec(carQuery, params...)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return -1, fmt.Errorf("%s: %w", op, errors.New("car not found"))
	}

	return 0, nil
}

func (s *Storage) patchOwner(carID int, patchOwner car.PatchPeople) (int, error) {
	const op = "storage.postgresql.PatchOwner"
	ownerQuery := "UPDATE PEOPLES SET "

	var params []interface{}
	var sql []string
	if patchOwner.Name.Valid {
		sql = append(sql, fmt.Sprintf("name = $%d,", len(sql)+1))
		params = append(params, patchOwner.Name.String)
	}
	if patchOwner.Surname.Valid {
		sql = append(sql, fmt.Sprintf("surname = $%d,", len(sql)+1))
		params = append(params, patchOwner.Surname.String)
	}
	if patchOwner.Patronymic.Valid {
		sql = append(sql, fmt.Sprintf("patronymic = $%d,", len(sql)+1))
		params = append(params, patchOwner.Patronymic.String)
	}

	if len(sql) == 0 {
		return 2, nil
	}
	ownerQuery += strings.Join(sql, " ")

	ownerQuery = ownerQuery[:len(ownerQuery)-1] + fmt.Sprintf(" WHERE id = (SELECT owner_id FROM CARS WHERE id = $%d)", len(sql)+1)
	params = append(params, carID)

	result, err := s.db.Exec(ownerQuery, params...)
	if err != nil {
		return -1, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return -1, nil
	}

	return 0, nil

}
