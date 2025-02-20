-- 你的验证码在redis中的key
-- phone_code:login:137xxxxxx
local key = KEYS[1]
-- 验证次数，我们一个验证码，最多重复三次，这个记录了验证了几次
-- phone_code:login:137xxxxxx:cnt
local cntKey = key ..":cnt"
-- 验证码 123456
local val = ARGV[1]
-- 过期时间
local ttl = tonumber(redis.call('ttl', key))

if ttl > 0 and ttl > 540 then
    -- key 存在，但是没有过期时间
    -- 系统错误，手动设置了这个 key ，但是没有设置过期时间
    return -1
else 
    -- 设置验证码
    redis.call('set', key, val)
    -- 设置过期时间
    redis.call('expire', key, 600)
    -- 设置验证次数
    redis.call('set', cntKey, 3)
    -- 设置验证次数过期时间
    redis.call('expire', cntKey, 600)
    return 0
end