require("block")

wcBlock.gpio = {}

function wcBlock.gpio.set(id, gpio, value)
	try(
		function()
			pio.pin.setdir(pio.OUTPUT, gpio)
			pio.pin.setpull(pio.NOPULL, gpio)
		end,
	    function(where, line, err, message)
			wcBlock.blockError(id, err, message)
		end
	)

	pio.pin.setval(value, gpio)	
end

function wcBlock.gpio.get(id, gpio)
	try(
		function()
			pio.pin.setdir(pio.INPUT, gpio)
			pio.pin.setpull(pio.PULLUP, gpio)
		end,
	    function(where, line, err, message)
			wcBlock.blockError(id, err, message)
		end
	)
	
	return  pio.pin.getval(gpio)
end