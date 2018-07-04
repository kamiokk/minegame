var API = {
    host: "",
    pollSpace: 500,
    requestJSON: function(path, data, callback) {
        $.ajax({
            url: API.host + path,
            contentType: "application/json",
            data: JSON.stringify(data),
            type: "POST",
            dataType: "JSON",
            success: function (result) {
                callback.call(this, result);
            },
            error: function (XMLHttpRequest, textStatus, errorThrown) {
                console.log('ajax_request_error');
            }
        });
    },
    checkAccountAvailable : function(account, callback) {
        $.getJSON("/user/checkAccountAvailable?account=" + account, callback)
    },
    register:function(account,password,passwordConf,callback) {
        data = {
            account:account,
            password:password,
            pwd_conf:passwordConf
        }
        this.requestJSON("/user/register", data, callback)
    },
    login: function (account, password, callback) {
        data = {
            account: account,
            password: password
        }
        this.requestJSON("/user/login", data, callback)
    },
    logout: function (callback) {
        this.requestJSON("/user/logout", null, callback)
    },
    userInfo: function (callback) {
        this.requestJSON("/user/info", null, callback)
    },
    giveOut: function (room, mine, callback) {
        data = {
            room: parseInt(room),
            mine: parseInt(mine)
        }
        this.requestJSON("/game/giveOut", data, callback)
    },
    gain: function (id) {
        data = {
            id: parseInt(id)
        }
        this.requestJSON("/game/gain", data, callback)
    },
    poll: function (roomID,t) {
        roomID = parseInt(roomID)
        t = parseInt(t)
        if (t > (new Date()).getTime() / 1000) {
            setTimeout(API.poll, API.pollSpace, roomID, t)
            return
        }
        data = {
            room: roomID,
            t: t
        }
        this.requestJSON("/game/poll", data, function(result) {
            if (result.code == 1) {
                console.log(result)
                setTimeout(API.poll, API.pollSpace, result.roomID, result.t)
            } else {
                alert("房间正忙，请稍候再试！")
            }
        })
    }
}