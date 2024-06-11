## xkcd-searcher

---
### Description

This is a simple web application that allows you to search [xkcd.com](https://xkcd.com/) comics and view them.

The application is written in Go and consists of three services: `webserver`, `xkcdserver`, and `authserver`. The `webserver` is responsible for serving the web interface, the `xkcdserver` is responsible for fetching comics from the xkcd API, and the `authserver` is responsible for user authentication. 

App stores all **_xkcd_.com** comics in a _PostgreSQL_ database and uses inverted indexes stored in _Redis_ for search. Some simple limiters are implemented to prevent the abuse.


![Recording-2024-05-31-231010](https://github.com/makarkananov/yadro-microservices/assets/52353806/ba10b7a4-7b8d-4116-b1bd-90139bbf9ae1)

---
### Usage

`make all` runs the entire application via `docker-compose`.

By default:
- `webserver` will be available on`:8081`
- `xkcdserver` will be available on `:8080`
- `authserver` will be available on `:50051`.

Default credentials are:
```
Username: admin
Password: password
```
**NB.** Before using the application, you need to update the comics. Here are some examples of requests for convenience:

1. Getting JWT token
```
curl --location 'http://localhost:8080/login' \
--header 'Content-Type: application/json' \
--data '{
    "username": "admin",
    "password": "password"
}'
```

2. Updating comics
```
curl --location --request POST 'http://localhost:8080/update' \
--header 'Authorization: Bearer some_token'
```
---
### Architecture
Here is the current architecture of the application:
![Untitled - Frame 1](https://github.com/makarkananov/yadro-microservices/assets/52353806/ef16a796-50e8-4073-b571-586416d8ceaa)
