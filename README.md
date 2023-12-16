# Ozon URL Shortener

##  Описание

API написаный на Go, с использованием pq и go-cache для работы с PostgreSQL и in-memmory данными.

## Endpoints

### GET 

```
GET /
```

Принимает JSON тело в запросе по типу:

```json
{
    "url":"http://example.com/test"
}
```

У ссылки в запросе обязательно должен быть путь после домена иначе нечего сокращать.

После приёма запроса, генерируется рандомный сокращённый путь, и возвращается в ответе

```text
http://example.com/aTyfW_25zG
```

### POST

```
POST /
```

Принимает сокращённую ссылку в теле запроса и возвращает оригинальную. Структура тела такая же как и у GET запроса.

```json
{
    "url":"http://example.com/aTyfW_25zG"
}
```

Возвращает текст:

```
http://example.com/test
```

## Установка и Запуск

API можно развернуть самому добавив в переменные окружения все необходимые данные для подключения к PostgreSQL, или же использовать докер.

### Docker

Чтобы использовать докер достаточно клонировать себе репозиторий и из папки репозитория выполнить команду:

```bash
sudo docker compose up
```

После чего будут добавлены три контейнера, один для API, другой с PostgreSQL и последний с базой данных для тестирования.

Чтобы изменить переменные для работы PostgreSQL и/или ADDR сервера, достаточно изменить их в файле .env.

Для запуска API сервера с in-memmory параметром необходимо в Dockerfile в разделе services->web поменять команду ` ./bin/ozonshrt` на `./bin/ozonshrt -m`

### Build

Перед сборкой API в переменные окружения необходимо добавить следующие данные (если вы не собираетесь использовать PostgreSQL то достаточно добавить только API_ADDR):

```env
POSTGRES_DB="ozonshrt"
POSTGRES_USER="ozon"
POSTGRES_PASSWORD="ozon"
POSTGRES_SSLMODE="disable"
POSTGRES_HOST="db"
POSTGRES_PORT="5432"
API_ADDR=":8080"
```

Для того чтобы самому скомпилировать API необходим go версии 1.21.5 и новее. Из папки с репозиторием запустить команду:

```shell
go build -o /bin/ozonshrt ./cmd/api
```

После чего появится папка с bin с исполняемым файлом ozonshrt. Чтобы запустить его нужно выполнить команду из папки с репозиторием 

```shell
./bin/ozonshrt
```

И если вы хотите использовать in-memmory хранилище то запускать сервер нужно с параметром -m

```shell
./bin/ozonshrt -m
```

## Тестирование

Чтобы протестировать сервер нужно либо подключится к контейнеру докера используя команду:

```shell
sudo docker exec -it ozon-web-1 bash
```

где `ozon-web-1` название контейнера с веб сервером, и ввести следующую команду, или же, если вы не используете докер, то из папки с репозиторием выполнить команду:

```shell
go test ./cmd/api
```

## Результаты стресс тестов

Используя k6 от Grafana, результаты стресс тестов следующие:

**GET запросы**

![](https://imgur.com/54urWTU.png)

**POST запросы**

![](https://imgur.com/JHHMcCm.png)

Мой JavaScript для POST запросов:

```javascript
import http from "k6/http";
import { SharedArray } from 'k6/data';

const data = new SharedArray('urls', function () {

  const f = JSON.parse(open('post.json'));
  return f; 
});


export default function () {
  const randomUrl = data[Math.floor(Math.random() * data.length)]
  const url = 'http://localhost:8080/';
  
  const payload = JSON.stringify({
    url: randomUrl.url
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  http.post(url, payload, params);
}
```

Скрипт для GET запросов:

```javascript
import http from "k6/http";
import { SharedArray } from 'k6/data';

const data = new SharedArray('urls', function () {

  const f = JSON.parse(open('get.json'));
  return f; 
});


export default function () {
  const randomUrl = data[Math.floor(Math.random() * data.length)]
  const url = 'http://localhost:8080/';
  
  const payload = JSON.stringify({
    url:randomUrl.url
  })

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  http.request("GET",url,payload,params);

}
```

Оба скрипта требуют файлы с URL'ами в формате массива где каждый элемент подходит под требования запроса.

