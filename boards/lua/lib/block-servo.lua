require("block")

wcBlock.servo = {}

function wcBlock.servo.attach(instance, gpio)
	if (_G[instance] == nil) then
		_G[instance] = servo.attach(gpio)
	end
end

function wcBlock.servo.move(id, gpio, value)
	local instance = "_servo"..gpio
	
	wcBlock.servo.attach(instance, gpio)
	
	_G[instance]:write(value)
end