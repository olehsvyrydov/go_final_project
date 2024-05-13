# Файлы для итогового задания

В директории `tests` находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.

Директория `web` содержит файлы фронтенда.

# Запуск тестов
Запустить приложение: go run .
В другом окне терминала запустить тесты: go test ./tests

# Docker
Собрать докер образ:
docker build --tag olehsvyrydov/final:latest .
docker run --rm -p 7540:7540 --name go_final olehsvyrydov/final:latest

# На заметку
Поменять порт можно указав в ***-e TODO_PORT=<port_number>*** и изменив значение параметра ***-p*** на соответствующий порт в docker run конструкции, описанной выше
Не реализована аутентификация