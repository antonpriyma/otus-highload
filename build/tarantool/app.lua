box.cfg{
    listen = 3301,
}
box.schema.user.passwd('pass')

local uuid = require('uuid')
local json = require('json')

users = box.schema.create_space('users', { if_not_exists = true })
users:format({
    {name="uuid",type="uuid"},
    {name="username",type="string"},
    {name="first_name",type="string"},
    {name="second_name",type="string"},
    {name="age",type="integer"},
    {name="sex",type="integer"},
    {name="city",type="string"},
    {name="biography",type="string"},
    {name="password",type="string"},
})

users:create_index('primary', {
    type = 'hash',
    parts = {1, 'uuid'},
    if_not_exists = true,
})

users:create_index('names', {
    type = 'tree',
    parts = {{3, 'string'},{4, 'string'}},
    if_not_exists = true,
})

function CreateUser(req)
    local user = {}
    -- set the user's properties based on the request
    user.uuid = req.uuid
    user.username = req.username
    user.first_name = req.first_name
    user.second_name = req.second_name
    user.age = req.age
    user.sex = req.sex
    user.city = req.city
    user.biography = req.biography
    user.password = req.password

    -- insert the user record into the "users" space
    users:insert({uuid.fromstr(req.uuid), req.username, req.first_name, req.second_name, req.age, req.sex, req.city, req.biography, req.password})

    -- return the created user record
    return user
end

function GetUser(req)
    user = users:get({uuid.fromstr(req.uuid)})
    return json.encode(user)
end

function FindUser(req)
    name = req.first_name
    surname = req.second_name
    user = users.index.names:get({name, surname})

    return json.encode(user)
end

return {
    CreateUser = CreateUser,
    GetUser = GetUser,
    FindUser = FindUser,
}