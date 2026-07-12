interface CriticalItemProp {
	label: string;
	info: string;
	unit?: string;
	arrow: string;
}

function CriticalItem({label, info, unit, arrow}: CriticalItemProp) {
	return (
		<div className="flex flex-col justify-center items-center gap-1 py-4 px-6 rounded-md bg-blue-400">
			<p className="text-base text-gray-600 leading-none">{label}</p>
			<p className="text-2xl whitespace-nowrap leading-none">{info}<span className="text-[16px]">{unit ? " " + unit : null}</span></p>
			<p className="text-3xl text-bold leading-none">{arrow}</p>
		</div>
	)
}

export default CriticalItem;
