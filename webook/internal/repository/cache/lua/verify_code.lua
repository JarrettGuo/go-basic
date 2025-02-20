local key = KEYS[1]
-- 用户输入的验证码
local expectedCode = ARGV[1]
local cntKey = key..":cnt"
local cnt = tonumber(redis.call('get', cntKey))
local code = redis.call('get', key)
if cnt == nil or cnt <= 0 then
    -- 验证次数用完
    return -1
elseif expectedCode == code then
    -- 验证码正确
    -- 验证次数清零
    redis.call("set", cntKey, 0)
    return 0
else
    -- 验证码错误
    -- 验证次数减一
    redis.call("decr", cntKey)
    return -2
end