# miniserver_go
2 Endpoints: GET, POST
basic URL "http://localhost:8080/"

++++++++++
+EXAMPLE:+
++++++++++


            ***************
            *Endpoint POST*
            ***************

request
_____________________________________________________________

POST / HTTP/1.1
Host: localhost:8080
Content-Type: text/plain

https://practicum.yandex.ru/ 
_____________________________________________________________


reply
_____________________________________________________________
HTTP/1.1 201 Created
Content-Type: text/plain
Content-Length: 30

http://localhost:8080/EwHXdJfB 
_____________________________________________________________


*************************************************************
*************************************************************

            ***************
            *Endpoint GET*
            ***************
request
_____________________________________________________________
GET /EwHXdJfB HTTP/1.1
Host: localhost:8080
Content-Type: text/plain 
_____________________________________________________________

reply
_____________________________________________________________
HTTP/1.1 307 Temporary Redirect
Location: https://practicum.yandex.ru/ 
_____________________________________________________________


*************************************************************
*************************************************************

go get go.uber.org/mock/gomock@latest
go get "github.com/stretchr/testify/assert"