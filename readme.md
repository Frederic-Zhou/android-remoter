# 备注

## 前提条件

- 可以打开ADB root 调试
- 程序写入到/system/xbin/AR目录
- 在init.**.rc里写入服务并在 on property:sys.boot_completed=1 里 start

## 编译 arm 版本

`GOOS=linux GOARCH=arm GOARM=7 go build`

## 将系统改为可写状态并进入

`adb root && adb shell mount -o rw,remount / && adb shell`

## init.**.rc

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

## 开发工具

编译调试
`adb root && adb shell mount -o rw,remount / && GOOS=linux GOARCH=arm GOARM=7 go build && adb push ./android-remoter /system/xbin/AR && adb shell reboot`


## useage

1. 在main.go 目录运行
`adb root && adb shell mount -o rw,remount / && GOOS=linux GOARCH=arm GOARM=7 go build && adb push ./android-remoter /system/xbin/AR && adb push ./assets /system/xbin/AR`

2. `adb shell` 进入手机linux ，修改 根目录下的 init.**.rc，添加内容如下

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
