# marketplace
simple marketplace


## Описание
Приложение реализует REST API для маркетплейса с регистрацией и авторизацией пользователей, размещением и просмотром объявлений с фильтрацией, сортировкой и пагинацией. Всё упаковано в Docker и написано на Go с использованием JWT для авторизации.

## **[<u>Задание</u>](docs/task.md)**

## Начало работы
### Установка
Клонирование репозитория
```sh
git clone https://github.com/AugustSerenity/marketplace
```
### Запуск сервиса
Запускаем контейнер с помощью Makefile
```sh
make run
```

Запускаем тесты с помощью Makefile
```sh
make text
```

## Пример работы с API

### 1. Регистрация пользователей

- **Регистрация Petr**:

Для регистрации пользователя используйте следующий `curl` запрос:

```bash
curl -X POST "http://localhost:8080/auth-register" \
   -H "Content-Type: application/json" \
   -d '{"login": "Petr", "password": "1234567"}'
```

- **Регистрация Pavel**:

```bash
curl -X POST "http://localhost:8080/auth-register" \
   -H "Content-Type: application/json" \
   -d '{"login": "Pavel", "password": "12345344"}'
```
### 2. Получение токенов
- **Токен для Petr**:

```bash
curl -X POST "http://localhost:8080/auth-login" \
   -H "Content-Type: application/json" \
   -d '{"login": "Petr", "password": "1234567"}'
```
-**Результат будет вида**:
```bash
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```
-**Скопируйте значение токена и сохраните его в переменную**:
```bash
PetrToken="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  # замените на ваш
```
- **Токен для Pavel**:
```bash
curl -X POST "http://localhost:8080/auth-login" \
   -H "Content-Type: application/json" \
   -d '{"login": "Pavel", "password": "12345344"}'
```
-**Результат будет вида**:
```bash
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```
-**Скопируйте значение токена и сохраните его в переменную**:
```bash
PavelToken="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  # замените на ваш
```

### 3. Создание объявлений
- **Petr создает объявления**:
```bash
curl -X POST "http://localhost:8080/create-ads" \
   -H "Authorization: Bearer $PetrToken" \
   -H "Content-Type: application/json" \
   -d '{"title": "Go for Beginners", "description": "A book about the Go programming language for beginners", "image_url": "http://img.com/go-book.jpg", "price": 25, "category": "Books"}'
```
```bash
curl -X POST "http://localhost:8080/create-ads" \
   -H "Authorization: Bearer $PetrToken" \
   -H "Content-Type: application/json" \
   -d '{"title": "Go The Quest for Solutions", "description": "Advanced guide for experienced Go developers", "image_url": "http://img.com/go-advanced-book.jpg", "price": 40, "category": "Books"}'
```
- **Pavel создает объявления**:
```bash
curl -X POST "http://localhost:8080/create-ads" \
   -H "Authorization: Bearer $PavelToken" \
   -H "Content-Type: application/json" \
   -d '{"title": "Fishing Rod", "description": "High-quality rod for professionals", "image_url": "http://img.com/fishing-rod.jpg", "price": 130, "category": "Fishing"}'
```
```bash
curl -X POST "http://localhost:8080/create-ads" \
   -H "Authorization: Bearer $PavelToken" \
   -H "Content-Type: application/json" \
   -d '{"title": "Fishing Kit", "description": "Complete set with line, hooks, and accessories", "image_url": "http://img.com/fishing-set.jpg", "price": 180, "category": "Fishing"}'
```

### 4. Просмотр объявлений
- **Показать все объявления**:
```bash
curl -X GET "http://localhost:8080/watch-ads"
```
- **Фильтрация по цене**:
```bash
curl -X GET "http://localhost:8080/watch-ads?min_price=100&max_price=200"
```
- **Сортировка по цене (по возрастанию)**:
```bash
curl -X GET "http://localhost:8080/watch-ads?sort_by=price&sort_order=asc"
```
- **Сортировка по дате создания (от самых новых)**:
```bash
curl -X GET "http://localhost:8080/watch-ads?sort_by=created_at&sort_order=desc"
```
- **Сортировка по дате создания (от самых старых)**:
```bash
curl -X GET "http://localhost:8080/watch-ads?sort_by=created_at&sort_order=asc"
```