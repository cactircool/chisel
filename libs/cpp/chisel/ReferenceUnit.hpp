#ifndef CHISEL_REFERENCE_UNIT_HPP
#define CHISEL_REFERENCE_UNIT_HPP

#include "BufferedConstructReader.hpp"
#include "BufferedData.hpp"
#include "ChiselIstream.hpp"
#include "Unit.hpp"
#include <cstdlib>

namespace chisel {

	class ReferenceUnit : public Unit {
		std::unique_ptr<BufferedConstructReader> _inner_construct;
		std::unique_ptr<Unit> _inner_unit;
	public:
		ReferenceUnit(chisel::ChiselIstream &reader) : Unit(reader) {}

		Unit::Type type() const override { return Unit::Type::REFERENCE; }
		BufferedData read() override {
			if (!_inner_construct)
				_inner_construct = extract_construct();
			if (!_inner_unit)
				_inner_unit = _inner_construct->next();
			auto data = _inner_unit->read();
			if (data.finished) {
				_inner_unit = _inner_construct->next();
				data = _inner_unit->read();
			}
			return data;
		}

		virtual std::unique_ptr<BufferedConstructReader> extract_construct() = 0;
	};

}

#endif
