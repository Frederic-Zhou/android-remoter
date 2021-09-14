import uiautomator2 as u2


def printinfo():
    d = u2.connect()  # connect to device
    print(d.info)
