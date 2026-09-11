[\[中文版\]](README_cn.md)

```
                _             _                 _       _          _         _             _
 ___ _ __  _ __(_)_ __   __ _| |__   ___   ___ | |_    (_)_      _| |_   ___| |_ __ _ _ __| |_ ___ _ __
/ __| '_ \| '__| | '_ \ / _` | '_ \ / _ \ / _ \| __|   | \ \ /\ / / __| / __| __/ _` | '__| __/ _ \ '__|
\__ \ |_) | |  | | | | | (_| | |_) | (_) | (_) | |_    | |\ V  V /| |_  \__ \ || (_| | |  | ||  __/ |
|___/ .__/|_|  |_|_| |_|\__, |_.__/ \___/ \___/ \__|  _/ | \_/\_/  \__| |___/\__\__,_|_|   \__\___|_|
    |_|                 |___/                        |__/
```

[![npm](https://img.shields.io/badge/demo-online-ed1c46.svg)](http://jwt.fanjin.io/)
[![Build Status](https://travis-ci.org/bfwg/springboot-jwt-starter.svg?branch=master)](https://travis-ci.org/bfwg/springboot-jwt-starter)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/bfwg/springboot-jwt-starter/blob/master/LICENSE)

> A Springboot token-based security starter kit featuring [AngularJS](https://angularjs.org/) and [Springboot](https://projects.spring.io/spring-boot/) ([JSON Web Token](https://jwt.io/))

> If you're looking for using Angular 4 for frontend implementation, please checkout [angular-spring-starter](https://github.com/bfwg/angular-spring-starter), a fullstack starter kit featuring [Angular 4](https://angular.io), [Router](https://angular.io/docs/ts/latest/guide/router.html), [Forms](https://angular.io/docs/ts/latest/guide/forms.html),
[Http](https://angular.io/docs/ts/latest/guide/server-communication.html),
[Services](https://gist.github.com/gdi2290/634101fec1671ee12b3e#_follow_@AngularClass_on_twitter),
[Spring boot](https://projects.spring.io/spring-boot/),
[Json Web Token](https://jwt.io/)


### [Live Demo](http://jwt.fanjin.io)
<p align="center">
    <img width="800" alt="Springboot JWT Starter" src="https://cloud.githubusercontent.com/assets/12819525/24693784/23c8af14-1994-11e7-9984-ebf612f740ec.png">
</p>

> Authentication is the most common scenario for using JWT. Once the user is logged in, each subsequent request will include the JWT, allowing the user to access routes, services, and resources that are permitted with that token. Single Sign On is a feature that widely uses JWT nowadays, because of its small overhead and its ability to be easily used across different domains.

> -- <cite>Auth0</cite>


### Quick start
**Make sure you have Go 1.27 or greater**

```bash
# clone our repo
# --depth 1 removes all but one .git commit history
git clone --depth 1 https://github.com/bfwg/springboot-jwt-starter.git

# change directory to our repo
cd springboot-jwt-starter

# download the dependencies
go mod download

# start the server
go run .

# the app will be running on port 8080
# there are two built-in user accounts to demonstrate the differing levels of access to the endpoints:
# - User - user:123
# - Admin - admin:123
```

### Run with docker

```
docker-compose up --build -d
```


### File Structure
```
springboot-jwt-starter/
 ├──internal/                                                   * our source files
 │   ├──app
 │   │   └──app.go                                              * wires every component together
 │   ├──apptest
 │   │   └──apptest.go                                          * test wiring: a wired app and role-bearing principals
 │   ├──clock                                                   * the injectable time source
 │   │   ├──clock.go
 │   │   └──mock.go                                             * scripted clock for tests
 │   ├──config
 │   │   ├──config.go                                           * loads the settings below
 │   │   └──application.yml                                     * application variables are configured here
 │   ├──db
 │   │   ├──db.go                                               * schema creation and seed loading
 │   │   └──import.sql                                          * database query (table creation)
 │   ├──model
 │   │   ├──authority.go
 │   │   ├──role.go
 │   │   ├──timestamp.go                                        * the wire format for instants
 │   │   ├──token_state.go                                      * JWT model
 │   │   └──user.go                                             * our main User model
 │   ├──repository                                              * repositories folder for accessing database
 │   │   └──user_repository.go
 │   ├──rest                                                    * rest endpoint folder
 │   │   ├──authentication_controller.go                        * auth related REST controller, refresh token endpoint etc.
 │   │   ├──json.go                                             * response rendering helpers
 │   │   └──user_controller.go                                  * REST controller to handle User related requests
 │   ├──security                                                * Security related folder(JWT, filters)
 │   │   ├──auth
 │   │   │   ├──authentication.go                               * the authenticated principal carried through a request
 │   │   │   ├──entry_point.go                                  * handle auth failures, like invalid token etc.
 │   │   │   ├──filter.go                                       * the JWT token filter and the role guard
 │   │   │   └──request.go                                      * login request object, contains username and password
 │   │   ├──crypto
 │   │   │   └──password_encoder.go                             * BCrypt hashing and verification
 │   │   └──token_helper.go                                     * token helper
 │   ├──server
 │   │   └──server.go                                           * routes, guards and static assets
 │   └──service
 │       ├──user_details_service.go                             * resolves a username to a principal
 │       └──user_service.go                                     * account service
 ├──web/                                                        * static assets are served here (Angular and html templates)
 ├──main.go                                                     * Application main entry point
 └──go.mod                                                      * what the go tool uses to manage its dependencies
```
# Table of Contents
* [File Structure](#file-structure)
* [Configuration](#configuration)
* [JSON Web Token](#json-web-token)

### Configuration
- **internal/server/server.go**: The server-side route and authentication configurations.
- **internal/config/application.yml**: Application level properties i.e the token expire time, token secret etc.
- **JWT token TTL**: JWT Tokens are configured to expire after 10 minutes, you can get a new token by signing in again.
- **Using a different database**: This Starter kit is using an embedded in-memory SQLite database. If you want to connect to another database you have to change the driver and the connection string in *internal/db/db.go*. Here is an example for a MySQL DB:

```go
import _ "github.com/go-sql-driver/mysql"

database, err := sql.Open("mysql", "myUser:myPassword@tcp(localhost:3306)/myDatabase?parseTime=true")
```

*Hint: the schema in `internal/db/db.go` uses SQLite's `INTEGER PRIMARY KEY AUTOINCREMENT`. For other databases you have to change that to the server's own auto-increment syntax, for example `AUTO_INCREMENT` on MySQL or `SERIAL` on PostgreSQL.*

### JSON Web Token
> JSON Web Tokens are an open, industry standard RFC 7519 method for representing claims securely between two parties.
for more info, checkout https://jwt.io/

### Contributing
I'll accept pretty much everything so feel free to open a Pull-Request


This project is inspried by
- [Cerberus](https://github.com/brahalla/Cerberus)
- [jwt-spring-security-demo](https://github.com/szerhusenBC/jwt-spring-security-demo)

___

# License
 [MIT](/LICENSE)

