# QuePasa

> A (micro) web-application to make web-based [WhatsApp][0] bots easy to write.

Built on the library [go-whatsapp][1] by [Rhymen][2].

**Implemented features:**

* Verify a number with a QR code
* Persistence of account data and keys
* Exposes HTTP endpoints for:
  * sending messages
  * receiving messages
  * download attachments
  * set webhook for receiving messages 

**WARNING: This application has not been audited. It should not be regarded as
secure, use at your own risk.**

**This is a third-party effort, and is NOT in any affiliated with [WhatsApp][0].**

### Why ?
 
Angry, Angry ... WhatsApp keeps canceling our number.  
 
When you need to communicate over WhatsApp from a different service, for example,
[a help desk](http://zammad.org/) or other web-app, QuePasa provides a simple HTTP
API to do so.

QuePasa stores keys and WhatsApp account data in a postgres database. It does
not come with HTTPS out of the box. Your QuePasa API tokens essentially give
full access to your WhatsApp account (to the extent that QuePasa has
implemented WhatsApp features). Use with caution.

For HTTPS use Nginx.

## If are you looking for a NODE.JS Project

Take a look at
https://github.com/pedroslopez/whatsapp-web.js/pulls

Its a lot more complete tool to whatsapp unofficial api

## Join our community 
Matrix chat room #cdr-link-dev-support:matrix.org
https://app.element.io/#/room/#cdr-link-dev-support:matrix.org

## Usage

## Prerequisites Local Devlopment (recomneded)

* Mysql (Recommended)
* Golang (Version go1.14.15)

### *installing golang*

```bash
cd /usr/src

sudo wget https://golang.org/dl/go1.14.15.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.14.15.linux-amd64.tar.gz

#export the PATH
export PATH=$PATH:/usr/local/go/bin

```



**First Step**

  Clone the repo 

  ```bash

git clone https://github.com/sufficit/sufficit-quepasa-fork.git

  ```

**Second Step**

  Create Database and Users

```bash

sudo mysql

# create the user

mysql> CREATE USER 'quepasa'@'%'IDENTIFIED BY 'S0me_RaNdoM_T3*T';

# Granting Permition to the Quepasa User

mysql> GRANT ALL ON quepasa.* TO 'quepasa'@'%';

# Flushing the Privileges 

mysql> FLUSH PRIVILEGES;

# Create quepasa DataBase 

mysql> CREATE DATABASE quepasa;

# exit mysql 

mysql> exit

```

**third Step**

  Creating the Tables Required

  ```bash
# cd into the cloned reop

cd <git_clone_location>/src/migrations/

#below will create the relevent tables in the quepasa database for you

sudo mysql --database=quepasa < 1_create_tables.up.sql

  ```
**forth Step**

Creating the .env file

```bash
# this file contains all the environment varibles that the system needed do the changes that matches your deployment

#create the .env file in the below location

nano <git_clone_location>/src/.env

# content of the file should looklike this 

WEBAPIHOST=0.0.0.0 
WEBAPIPORT=31000 # web port of the API
WEBSOCKETSSL=false # http or Https
DBDRIVER=mysql #Databse Server
DBHOST=localhost
DBDATABASE=quepasa
DBPORT=3306
DBUSER=quepasa
DBPASSWORD='S0me_RaNdoM_T3*T' #the string you created in the third step 
DBSSLMODE=disable
APP_ENV=development # this will write some extra debug messages you can change it to production if needed
MIGRATIONS=false
SIGNING_SECRET=5345fgdgfd54asdasdasdd #some random test this will be used for password encription 

```

**fifth Step**

Compiling the Packge

```bash
# cd into the src directory

<git_clone_location>/src/

# compile using golang this may take few seconds to compile

go run main.go

```

**final step**

go to http://yourIP:3100/setup in the web browser and register an admin user for your system
    


if error occourd such as go not found please make sure to [export the path](#installing-golang)


### Prerequisites

For local development
* docker
* golang
* postgresql

### Run using Docker

* Add info about database migrations

```bash

make docker_build
# edit docker-compose.yml.sample to your hearts content
docker-compose up
```

## HTTP API

1. Use the `Accept: application/json` header
2. `TOKEN` should be treated like a password.

### Get bot info

A simple method for testing your bot's auth token. Requires no parameters. Returns basic information about the bot.

**request**
```
GET /bot/<TOKEN>/
```

***response***

```json
{
    "id": "5454544554343@c.us",
    "user_id": "845ae4d0-f2c3-5342-91a2-5b45cb8db57c",
    "token": "8129c0b4-0b96-4486-84fc-c3dd7b03f846",
	"webhook" : "",
    "is_verified": true,
    "created_at": "2018-11-02T11:36:24.273Z",
    "updated_at": "2018-11-02T11:36:24.273Z"
}

```

### Sending

**request**
```
POST /bot/<TOKEN>/send

{
  "recipient": "+15555555552",
  "messsage": "Hello World!"
}
```

**response**
```json
{
  "result": {
    "recipient": "+15555555551",
    "source": "+15555555552",
    "status": "sent",
    "timestamp": "1543420505142"
  }
}
```

### Receive

The "timestamp" query parameter is optional. A maximum of 40 messages per conversation will be returned.

**request**
```
GET /bot/<TOKEN>/receive?timestamp=1541265073783
```

**response**
```json
{
  "messages": [
    {
      "source": "+15555555551",
      "timestamp": "1541265073894",
      "message": {
        "body": "Hello World!",
        "profileKey": "XXTXQ=="
      }
    }
  ],
  "bot": {
    "id": "129f1757-e706-452e-aa1c-4994a95e1092",
    "number": "+15555555552",
    "user_id": "845ae4d0-f2c3-5342-91a2-5b45cb8db57c",
    "token": "8129c0b4-0b96-4486-84fc-c3dd7b03f846",
    "is_verified": true,
    "created_at": "2018-11-02T11:36:24.273Z",
    "updated_at": "2018-11-02T11:36:24.273Z"
  }
}
```
## Environment Variables

WEBAPIPORT:			31000				#
PGHOST: 			"localhost"			#
PGDATABASE:			"quepasa_dev"   	#
PGPORT:				"5432"				#
PGUSER:				"quepasa"			#
PGPASSWORD:			"quepasa"			#
PGSSLMODE:			"disable"			#
APP_ENV:			"development"		#
DEBUGREQUESTS:		true				#
DEBUGJSONMESSAGES:	true				#
SIGNING_SECRET:		"any secret here"	#
TZ:					"America/Sao_Paulo"	#

## License

[![License GNU AGPL v3.0](https://img.shields.io/badge/License-AGPL%203.0-lightgrey.svg)](https://github.com/sufficit/sufficit-quepasa-fork/blob/master/LICENSE.md)

QuePasa is a free software project licensed under the GNU Affero General Public License v3.0 (GNU AGPLv3) by [Sufficit Soluções](https://www.sufficit.com.br).

[0]: https://whatsapp.com
[1]: https://github.com/Rhymen/go-whatsapp
[2]: https://github.com/Rhymen
