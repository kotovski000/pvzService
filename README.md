# Управление ПВЗ (Пункты Выдачи Заказов)

## Установка

```
git clone https://github.com/ваш-репозиторий/pvzService.git
cd pvzService
go mod tidy
```

## Настройка

Приложение использует несколько переменных окружения:

- ```DATABASE_HOST```: Хост базы данных. По умолчанию используется db.  
- ```DATABASE_PORT```: Порт базы данных. По умолчанию используется 5432.  
- ```DATABASE_USER```: Имя пользователя для подключения к базе данных. По умолчанию используется postgres.  
- ```DATABASE_PASSWORD```: Пароль для подключения к базе данных. Установите его на значение, которое вы используете (например, password).  
- ```DATABASE_NAME```: Имя базы данных. По умолчанию используется pvz.  
- ```SERVER_PORT```: Порт, на котором будет работать сервер. По умолчанию используется порт 8080.  
- ```JWT_SECRET```: Секретный ключ для аутентификации JWT. Установите его на значение, которое вы хотите использовать (например, your-secret-key).  

## Структура проекта
```
.
├── cmd/app/                  # Основное приложение
│   ├── main.go               # Точка входа
│   └── makeApp.go            # Инициализация приложения
├── internal/                 # Внутренние модули
│   ├── config/               # Конфигурация
│   ├── db/                   # Подключение к БД
│   ├── grpc/                 # gRPC сервер
│   ├── handlers/             # HTTP обработчики
│   ├── middleware/           # Промежуточное ПО
│   ├── models/               # Модели данных
│   ├── processors/           # Бизнес-логика
│   ├── prometheus/           # Метрики Prometheus
│   ├── proto/                # Protobuf файлы
│   ├── repository/           # Работа с БД
│   ├── tests/                # Интеграционные тесты
│   └── utils/                # Вспомогательные утилиты
├── migrations/               # Миграции БД
├── taskСondition/            # Условия задачи 
├── .env                      # Переменные окружения
├── Dockerfile                # Конфигурация Docker
├── docker-compose.yaml       # Конфигурация Docker Compose
├── go.mod                    # Зависимости Go
├── go.sum                    # Хеши зависимостей
└── prometheus.yml            # Конфигурация Prometheus
```

## Мониторинг
- Prometheus доступен на ```http://localhost:9090```;
- Метрики приложения доступны на ```http://localhost:<порт-метрики>/metrics```;

## GRPC
- GRPC доступен на ```http://localhost:3000```
- Возвращает все добавленные в систему ПВЗ.

## Тестирование
- Unit-тесты запускаются через Dockerfile;
- После успешного прохождения тестов, собирается образ;
- Интеграционный тест запускается после того, как запуститься всё приложение через консоль в корне проекта следующей командой ``` go test ./internal/tests/... -v```;
- Вывод процента тестового покрытия можно осуществить через консоль в корне проекта следующей командой ```go test -cover ./internal/handlers/... ./internal/processors/... ./internal/repository/...```;
- Процент покрытия:
``` 
ok      pvzService/internal/handlers    0.408s  coverage: 85.3% of statements
ok      pvzService/internal/processors  11.306s coverage: 81.5% of statements
ok      pvzService/internal/repository  0.599s  coverage: 84.5% of statements
```

