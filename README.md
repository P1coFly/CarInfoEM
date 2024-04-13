# Реализовать каталог автомобилей.

## Необходимо реализовать следующее:
1. Выставить rest методы
	1. Получение данных с фильтрацией по всем полям и пагинацией 
	2. Удаления по идентификатору
	3. Изменение одного или нескольких полей по идентификатору
	4. Добавления новых автомобилей в формате
```json
{
    "regNums": ["X123XX150"] // массив гос. номеров
}
```
2. При добавлении сделать запрос во внешнее АПИ, описанного сваггером (это описание некоторого внешнего АПИ, которого нет, но к которому надо обращаться. Реализованное, согласно описанию, АПИ будет использоваться при проверке)

```yaml
openapi: 3.0.3
info:
  title: Car info
  version: 0.0.1
paths:
  /info:
    get:
      parameters:
        - name: regNum
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Car'
        '400':
          description: Bad request
        '500':
          description: Internal server error
components:
  schemas:
    Car:
      required:
        - regNum
        - mark
        - model
        - owner
      type: object
      properties:
        regNum:
          type: string
          example: X123XX150
        mark:
          type: string
          example: Lada
        model:
          type: string
          example: Vesta
        year:
          type: integer
          example: 2002
        owner:
          $ref: '#/components/schemas/People'
    People:
      required:
        - name
        - surname
      type: object
      properties:
        name:
          type: string
        surname:
          type: string
        patronymic:
          type: string
```
3. Обогащенную информацию положить в БД postgres (структура БД должна быть создана путем миграций при старте сервиса)
4. Покрыть код debug- и info-логами
5. Вынести конфигурационные данные в .env-файл
6. Сгенерировать сваггер на реализованное АПИ


# Запуск проекта

## Запуск через Docker-контейнер.
Необходимо ввести в терминал из корня проекта - ```docker-compose up --build -d```\
Будет развёрнуто 2 контейнера:
1. db - СУБД PostgreSQL
2. server - сам сервис 

## Запуск локально
### Конфигурация БД
Сервис использует Postgresql, для успешной работы сервиса подтребуется:
1. Создать бд VK_Authorization
2. Воспользоваться скриптом ```internal\postgresql\init.sql```\

После использования скрипта:
- Будет создан и настроен пользователь auth_service
- Будет создана таблица USERS

### Запуск сервиса
1. Необходимо установить зависимости, для это из корня проекта надо выполнить команду - ```go mod download```
2. Скомпелировать выполняемый файл -  ```go build -o <название_выполняемого_файла> ./cmd/auth-server/main.go```
4. Установить переменные окружения ```CONFIG_PATH=./config/config.yml``` и ```KEY_JWT=testkey```
3. Запуск - ```./app```

# Проверка

Для удобной проверки была сгенирирована спецификация swagger (использовался подход code-first). Спецификация находится в директории docs. Также воспользоваться спецификацией можно по URI - /swagger/ (например: http://localhost:8080/swagger/)

# Конфигурация

Для конфигурации проекта надо изменить файл ./config/config.yml
Также по необходимости dockerfile и dockercompose

# Доступ к /feed

Для получения доступа к /feed, необходимо вставить полученный токен доступа в заголовок запроса:
```Authorization: Bearer <access_token>```