#ifndef CHISEL_OR_UNIT_HPP
#define CHISEL_OR_UNIT_HPP

#include <memory>
#include "Unit.hpp"

namespace chisel {

	class OrUnit : public chisel::Unit {
		std::unique_ptr<Unit> _inner_unit;
	public:
		OrUnit(chisel::ChiselIstream &reader) : chisel::Unit(reader) {}

		Unit::Type type() const override { return Unit::Type::OR; }
		chisel::BufferedData read() override {
			if (!_inner_unit)
				_inner_unit = extract_option();
			return _inner_unit->read();
		}
		virtual std::unique_ptr<Unit> extract_option() = 0;
	};

}

#endif
