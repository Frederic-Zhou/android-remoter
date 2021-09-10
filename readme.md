# Readme

## 前提条件

- 可以打开ADB root 调试
- 程序写入到/system/xbin/AR目录
- 在init.**.rc里写入服务并在 on property:sys.boot_completed=1 里 start

## 编译 arm 版本

`GOOS=linux GOARCH=arm GOARM=7 go build`

## 将系统改为可写状态并进入

`adb root && adb remount / && adb shell`

## init.**.rc

```conf
service androidremoter /system/xbin/AR/android-remoter
    user root
    group root
    disabled
    onshot
    seclabel u:r:su:s0

on property:dev.bootcomplete=1
   setprop service.adb.tcp.port 5555
   stop adbd
   start adbd

on property:sys.boot_completed=1
   stop androidremoter
   start androidremoter
```

## useage

1. 在main.go 目录运行
`adb root && adb remount && adb shell mkdir -p /system/xbin/AR && GOOS=linux GOARCH=arm GOARM=7 go build && adb push ./android-remoter /system/xbin/AR && adb push ./assets /system/xbin/AR`



1. `adb shell` 进入手机linux ，修改 根目录下的 init.**.rc，添加内容如下

``` sh
service androidremoter /system/xbin/AR/android-remoter
    user root
    group root
    disabled
    onshot
    seclabel u:r:su:s0

on property:dev.bootcomplete=1
   setprop service.adb.tcp.port 5555
   stop adbd
   start adbd

on property:sys.boot_completed=1
   stop androidremoter
   start androidremoter
```

3. 重启手机

4. 手机启动后，当连入互联网后，自动启动atx-agent,term(ttyd),frpc 并且frpc会将用到的端口转发到服务器

5. 通过访问127.0.0.1:8000/setfrpc 设置 atx term frpc 接口
   - 8000 主程序端口
   - 7912 atx端口
   - 8100 term端口
   - 8200 frpc管理端口
   - frps服务器地址和端口（与运行frps的服务器必须一致） frps与frpc通信的地址和端口

```shell
curl --location --request POST 'http://<手机IP地址>:8000/setfrpc' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'serverAddr=<服务器IP地址>' \
--data-urlencode 'serverPort=7000' \
--data-urlencode 'devicesID=testxxx-<可以加上DeviceID>' \
--data-urlencode 'user=admin' \
--data-urlencode 'pwd=123'
```

6. 运行FRPS服务器
`./frps -c ./frps.ini`
配置文件
```ini
[common]
bind_port = 7000
vhost_http_port = 8080

dashboard_port = 7555
dashboard_user = admin
dashboard_pwd = admin

subdomain_host = localhost
```

## 问题解决
有一些手机atx-agent的投屏和触控不可用，因为atx-agent里没有正确的下载对应的minicap.so
可以到minicap官方 下载 https://github.com/DeviceFarmer/minicap
还可以简单的在连接上adb后，用`python -m uiautomator2 init`初始化一次，安装上正确的版本

## 相关资料 

frp： https://github.com/fatedier/frp
atx-agent: https://github.com/openatx/atx-agent
minicap: https://github.com/DeviceFarmer/minicap
uiautomator2: https://github.com/openatx/uiautomator2
ttyd: https://github.com/tsl0922/ttyd