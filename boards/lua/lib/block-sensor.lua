require("block")

wcBlock.sensor = {}

function wcBlock.sensor.attach(instance, type, ...)
	if (_G[instance] == nil) then
		_G[instance] = sensor.setup(type, ...)
	end
end

function wcBlock.sensor.read(id, sid, type, magnitude, ...)
	local instance = "_"..sid.."_"..type
	
	wcBlock.sensor.attach(instance, type, ...)
	
	return _G[instance]:read(magnitude)
end

function wcBlock.sensor.set(id, sid, type, setting, value, ...)
	local instance = "_"..sid.."_"..type
	
	wcBlock.sensor.attach(instance, type, ...)
	
	return _G[instance]:set(setting, value)
end