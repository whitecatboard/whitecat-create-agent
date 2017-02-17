require("block")

wcBlock.adc = {}

function wcBlock.adc.get(id, aid, channel)
	local instance = "_"..aid
	
	if (_G["_adc1"] == nil) then
	    _G["_adc1"] = adc.setup(adc.ADC1)
	end

	if (_G[instance] == nil) then
		_G[instance] = _adc1:setupchan(12, channel)
	end
	
	return _G[instance]:read()
end