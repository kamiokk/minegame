const MsgNotEnoughPoint = "金币不足，请先充值"
const MsgError = "系统正忙，请稍后再试"

var API = {
    pollSpace: 500,
    pollCallback: function(){},
    getHost: function(){
        return location.protocol + "//" + location.host
    },
    requestJSON: function(path, data, callback) {
        $.ajax({
            url: API.getHost() + path,
            contentType: "application/json",
            data: JSON.stringify(data),
            type: "POST",
            dataType: "JSON",
            success: function (result) {
                callback.call(this, result)
            },
            error: function (XMLHttpRequest, textStatus, errorThrown) {
                ret = {
                    "code": -1,
                    "msg": "request error",
                    "textStatus": textStatus
                }
                callback.call(this, ret)
            }
        });
    },
    checkAccountAvailable : function(account, callback) {
        $.getJSON("/user/checkAccountAvailable?account=" + account, callback)
    },
    register:function(account,password,passwordConf,agent,callback) {
        data = {
            account: account,
            password: password,
            pwd_conf: passwordConf,
            agent:agent
        }
        API.requestJSON("/user/register", data, callback)
    },
    isLogin: function (callback) {
        $.getJSON("/user/isLogin", callback)
    },
    login: function (account, password, callback) {
        data = {
            account: account,
            password: password
        }
        API.requestJSON("/user/login", data, callback)
    },
    logout: function (callback) {
        API.requestJSON("/user/logout", null, callback)
    },
    userInfo: function (callback) {
        API.requestJSON("/user/info", null, callback)
    },
    agentCount: function (callback) {
        API.requestJSON("/user/agentCount", null, callback)
    },
    balanceLog: function (offset,callback) {
        $.getJSON("/user/balanceLog?offset=" + offset, callback)
    },
    userStat: function (callback) {
        $.getJSON("/user/stat", callback)
    },
    giveOut: function (room, mine, callback) {
        data = {
            room: parseInt(room),
            mine: parseInt(mine)
        }
        API.requestJSON("/game/giveOut", data, callback)
    },
    gain: function (id, callback) {
        data = {
            id: parseInt(id)
        }
        API.requestJSON("/game/gain", data, callback)
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
        API.requestJSON("/game/poll", data, function(result) {
            API.pollCallback.call(this,result)
            if (result.code == 1) {
                setTimeout(API.poll, API.pollSpace, result.roomID, result.t)
            }
        })
    }
};

var Utils = {
    validator: {
        account: function (val) {
            var reg = /^[A-Za-z0-9]{6,12}$/
            return reg.test(val)
        },
        password: function (val) {
            var reg = /^[A-Za-z0-9]{6,12}$/
            return reg.test(val)
        }
    },
    f2y: function(val) {
        return parseFloat(val) / 100
    },
    y2f: function(val) {
        return parseInt(val * 100)
    },
    getQueryVariable: function(variable)
    {
        var query = window.location.search.substring(1);
        var vars = query.split("&");
        for (var i = 0; i < vars.length; i++) {
        var pair = vars[i].split("=");
            if (pair[0] == variable) { return pair[1]; }
        }
        return "";
    }
};