require("block")

wcBlock.gpio = {}

function wcBlock.gpio.set(id, gpio, value)
	pio.pin.setdir(pio.OUTPUT, gpio)
	pio.pin.setpull(pio.NOPULL, gpio)
	pio.pin.setval(value, gpio)	
end

function wcBlock.gpio.get(id, gpio, value)
	pio.pin.setdir(pio.INPUT, gpio)
	pio.pin.setpull(pio.PULLUP, gpio)
	
	return pio.pin.getval(gpio)
end