package adder

import (
	"log/slog"
	"net/http"

	"github.com/P1coFly/CarInfoEM/http-server/handlers/err_response"
	"github.com/P1coFly/CarInfoEM/internal/models/car"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	RegNums []string `json:"reg_num" example:"X123XX150"`
} //@name RegNums

type AddCar interface {
	AddCar(car car.Car) (int, error)
}

type CarInfo interface {
	Get(regNum string) (car.Car, int, error)
}

type AddResponse struct {
	FailedCars []string `json:"failed_cars,omitempty"`
	Errors     []error  `json:"errors,omitempty"`
	CarsID     []int    `json:"cars_id,omitempty"`
}

// @Summary Add
// @Tags car
// @Description add car
// @Accept json
// @Produce json
// @Param input body RegNums true "Array of new car registration numbers"
// @Success 201,206 {object} AddResponse
// @Failure 400 {object} err_response.Response
// @Failure 500 {object} err_response.Response
// @Failure default {object} err_response.Response
// @Router /car/add [post]
func New(log *slog.Logger, adder AddCar, carInfo CarInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.AddCar.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		//получаем дату из запроса
		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", "error", err)
			w.WriteHeader(400)
			render.JSON(w, r, err_response.Error("failed to decode request body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		/*отправляем regNum в carInfo
		получаем объект car и записываем в бд
		в случаи ошибки запоминаем её и переходим к следующему regNum*/
		var failedCars []string
		var errors []error
		var successfulCarIDs []int
		var code int
		for _, regNum := range req.RegNums {
			car, temp, err := carInfo.Get(regNum)
			code = temp
			if err != nil {
				failedCars = append(failedCars, regNum)
				errors = append(errors, err)
				continue
			}

			carID, err := adder.AddCar(car)
			if err != nil {
				failedCars = append(failedCars, regNum)
				errors = append(errors, err)
				continue
			}
			successfulCarIDs = append(successfulCarIDs, carID)
		}

		//Если со всеми regNum случилась ошибка
		if len(failedCars) == len(req.RegNums) {
			w.WriteHeader(code)
			render.JSON(w, r, err_response.Error("failed to add cars"))
			return
		}
		//Если частично с regNum случилась ошибка
		if len(failedCars) > 0 {
			w.WriteHeader(206)
			render.JSON(w, r, AddResponse{
				FailedCars: failedCars,
				Errors:     errors,
				CarsID:     successfulCarIDs,
			})
			return
		}

		w.WriteHeader(201)
		render.JSON(w, r, AddResponse{
			CarsID: successfulCarIDs,
		})
	}
}
