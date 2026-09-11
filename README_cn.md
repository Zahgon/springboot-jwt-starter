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

> 这是一个启动项目用到[Springboot](https://projects.spring.io/spring-boot/)，[AngularJS](https://angularjs.org/)和([JSON Web Token](https://jwt.io/)).


> 如果你想用Angular4+来做前端，请看这个项目：[angular-spring-starter](https://github.com/bfwg/angular-spring-starter), 这是一个全栈启动项目用到[Angular 4](https://angular.io), [Router](https://angular.io/docs/ts/latest/guide/router.html), [Forms](https://angular.io/docs/ts/latest/guide/forms.html),

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


### 快速开始
**要用Go 1.27或更高哦**

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

# 后端服务会跑在端口 8080
# 有两个账户预设在内存:
# - User - user:123
# - Admin - admin:123
```


### 文件结构
```
springboot-jwt-starter/
 ├──internal/                                                   * 源代码
 │   ├──app
 │   │   └──app.go                                              * 把所有组件装配起来
 │   ├──apptest
 │   │   └──apptest.go                                          * 单元测试用的装配帮助库
 │   ├──clock                                                   * 可注入的时间源
 │   │   ├──clock.go
 │   │   └──mock.go                                             * 测试用的可编排时钟
 │   ├──config
 │   │   ├──config.go                                           * 读取下面的配置文件
 │   │   └──application.yml                                     * 程序配置文件
 │   ├──db
 │   │   ├──db.go                                               * 建表以及载入初始数据
 │   │   └──import.sql                                          * database query，当程序运转时，这里的用户数据会被load到内存
 │   ├──model
 │   │   ├──authority.go
 │   │   ├──role.go
 │   │   ├──timestamp.go                                        * 时间的输出格式
 │   │   ├──token_state.go                                      * 用来返还给用户的JSON object
 │   │   └──user.go                                             * 用户模型
 │   ├──repository                                              * 访问数据库
 │   │   └──user_repository.go
 │   ├──rest
 │   │   ├──authentication_controller.go                        * 用来身份认证，刷新token
 │   │   ├──json.go                                             * 响应输出帮助库
 │   │   └──user_controller.go                                  * 用来获取用户信息
 │   ├──security
 │   │   ├──auth
 │   │   │   ├──authentication.go                               * 用来帮助JWT身份认证
 │   │   │   ├──entry_point.go                                  * 用来应付登录失败的端点
 │   │   │   ├──filter.go                                       * **JWT过滤器**以及权限守卫
 │   │   │   └──request.go                                      * 登录请求模型
 │   │   ├──crypto
 │   │   │   └──password_encoder.go                             * BCrypt 加密与校验
 │   │   └──token_helper.go                                     * 识别Token的帮助库
 │   ├──server
 │   │   └──server.go                                           * 路由，权限守卫以及静态文件
 │   └──service
 │       ├──user_details_service.go                             * 用来核对用户登录信息
 │       └──user_service.go                                     * 用户服务
 ├──web/                                                        * 静态文件夹，所有AngularJS代码都在这里
 ├──main.go                                                     * 入口文件
 └──go.mod                                                      * Go 依赖配置文件
```
# Table of Contents
* [文件结构](#文件结构)
* [配置](#配置)

### 配置
- **internal/server/server.go**: 配置路由、登录过滤器以及其他安全相关的配置。
- **internal/config/application.yml**: 程序相关配置例如：token过期时间，token的秘密是什么。
- **JWT token TTL**: JWT默认配置是10分钟过期。
- **用别的数据库**: 这个项目是用内存SQLite数据库。如果你想用一个真正的数据库例如PostreSql，MySql，你需要改 **internal/db/db.go** 里的驱动跟连接串。例如说如果你要用MySQL：

```go
import _ "github.com/go-sql-driver/mysql"

database, err := sql.Open("mysql", "myUser:myPassword@tcp(localhost:3306)/myDatabase?parseTime=true")
```



___

# License
 [MIT](/LICENSE)

