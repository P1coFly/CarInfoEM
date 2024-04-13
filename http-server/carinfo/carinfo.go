package carinfo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/P1coFly/CarInfoEM/internal/models/car"
)

type CarInfoService struct {
	Host string
}

func New(host string) *CarInfoService {
	return &CarInfoService{Host: host}
}

func (c *CarInfoService) Get(regNum string) (car.Car, int, error) {
	url := fmt.Sprintf("%s/info?regNum=%s", c.Host, regNum)
	resp, err := http.Get(url)
	if err != nil {
		return car.Car{}, 500, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return car.Car{}, resp.StatusCode, fmt.Errorf("failed to get car information: status code %d", resp.StatusCode)
	}

	var carData car.Car
	err = json.NewDecoder(resp.Body).Decode(&carData)
	if err != nil {
		return car.Car{}, 500, err
	}

	return carData, resp.StatusCode, nil
}
