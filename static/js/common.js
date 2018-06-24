var API = {
    host: "",
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
    }
}