local headers = { }

headers['Content-Type'] = "application/x-www-form-urlencoded"

id_start = math.random(0, 100*1000)

request = function()
    id = math.random(id_start, id_start+200)
    body = "username=user-"..id.."&password=user-"..id
    api = "/login"
    return wrk.format("POST", api, headers, body)
end
