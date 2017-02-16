wcBlock = {
	delevepMode = true
}

function wcBlock.blockStart(id)
	if (wcBlock.delevepMode) then
		print("<blockStart,"..id..">\r\n")
	end
end

function wcBlock.blockEnd(id)
	if (wcBlock.delevepMode) then
		print("<blockEnd,"..id..">\r\n")
	end
end

function wcBlock.blockError(id)
	if (wcBlock.delevepMode) then
		print("<blockError,"..id..">\r\n")
	end
end