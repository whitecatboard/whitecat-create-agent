os.loglevel(os.LOG_ERR)

wcBlock = {
	delevepMode = false
}

function wcBlock.blockStart(id)
	if (wcBlock.delevepMode) then
		uart.lock(uart.CONSOLE)
		uart.write(uart.CONSOLE,"<blockStart,")
		uart.write(uart.CONSOLE,id)
		uart.write(uart.CONSOLE,">\n")
		uart.unlock(uart.CONSOLE)
	end
end

function wcBlock.blockEnd(id)
	if (wcBlock.delevepMode) then
		uart.lock(uart.CONSOLE)
		uart.write(uart.CONSOLE,"<blockEnd,")
		uart.write(uart.CONSOLE,id)
		uart.write(uart.CONSOLE,">\n")
		uart.unlock(uart.CONSOLE)
	end
end

function wcBlock.blockError(id, err, msg)
	if (wcBlock.delevepMode) then
		uart.lock(uart.CONSOLE)
		uart.write(uart.CONSOLE,"<blockError,")
		uart.write(uart.CONSOLE,id)
		uart.write(uart.CONSOLE,",")
		uart.write(uart.CONSOLE,msg)
		uart.write(uart.CONSOLE,">\n")
		uart.unlock(uart.CONSOLE)
	end
	
	error(err..":"..msg)
end