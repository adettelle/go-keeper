## yandex practicum final task

## 1. запуск minio 
зайти в админ панель, создать access ключ и создать бакет

docker-compose up -d minio

## запуск сервера 
переменные окружения: 
# по умолчанию:
ADDRESS='localhost:8080'
DATABASE_HOST='localhost'
DATABASE_PORT='5433'
DATABASE_USER='postgres'
DATABASE_PASSWORD='password'
DATABASE_NAME='praktikum-fin'

MINIO_ENDPOINT='localhost:9000'
USE_SSL='false'

# обязательно задать:
JWT_SIGNKEY
ACCESS_KEYID
SECRET_ACCESSKEY

go run ./cmd/server/ 

## регистрация клиента 

go run ./cmd/client register -n <name> -l <email> -p <masterpassword>

## логин клиента 

go run ./cmd/client login -l <email> -p <password>

## получение всех паролей клиента 

go run ./cmd/client passwords

## добавление файла клиента 

go run ./cmd/client add-file -p <path_to_file> -t <unique_title> -d <description> 

## получение (сохранение) файла клиента 

go run ./cmd/client get-file -i <file_id> > <path_to_file_to_save>

