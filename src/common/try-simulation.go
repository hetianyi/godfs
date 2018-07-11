package common

func Try(f func(), catcher func(interface{})) {
    defer func() {
        if err := recover(); err != nil {
            catcher(err)
        }
    }(); f()
}


