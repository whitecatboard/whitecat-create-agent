require("block")

wcBlock.adc = {}

function wcBlock.adc.get(id, aid, channel)
	local instance = "_"..aid
	
	if (_G[instance] == nil) then
	    _G[instance] = adc.setup(adc.ADC1, channel, 12)
	end

	return _G[instance]:read()
end