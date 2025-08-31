#ifndef CHISEL_UNIT_HPP
#define CHISEL_UNIT_HPP

#include <iostream>
#include <iterator>
#include "BufferedData.hpp"
#include "ChiselIstream.hpp"

namespace chisel {

	class Unit {
		chisel::ChiselIstream *_reader;
	public:
		enum Type {
			UNIT,
			REFERENCE,
			OR,
		};

		// Unit() : _reader(nullptr) {}
		Unit(chisel::ChiselIstream &reader) : _reader(&reader) {}
		virtual ~Unit() = default;

		// bool reader_set() const { return _reader; }
		void reader(chisel::ChiselIstream &reader) { _reader = &reader; }
		chisel::ChiselIstream &reader() { return *_reader; }
		virtual Type type() const { return UNIT; }
		chisel::ChiselIstream &reader() const { return *_reader; }

		virtual void reset() {}
		virtual BufferedData read() = 0;
		void skip(size_t n) { reader().seekg(n, std::ios::cur); }
		void unread() { reader().seekg(-1, std::ios::cur); }
		void unread(long long n) { reader().seekg(-n, std::ios::cur); }

		void skip() { for (auto _ : *this) {} }

		class Iterator {
			Unit *_unit_reader;
			BufferedData _state;

			Iterator(Unit &unit_reader, BufferedData state) : _unit_reader(&unit_reader), _state(state) {}

		public:
			using iterator_category = std::forward_iterator_tag;
	        using difference_type   = std::ptrdiff_t;
	        using value_type        = char;
	        using pointer           = char *;
	        using reference         = char &;

			Iterator(Unit &unit_reader) : _unit_reader(&unit_reader), _state(_unit_reader->read()) {}

			Iterator(const Iterator &other) : _unit_reader(other._unit_reader), _state(other._state) {}

			bool operator==(const Iterator &other) const {
				return _unit_reader == other._unit_reader && _state.data == other._state.data;
			}

			bool operator!=(const Iterator &other) const {
				return _unit_reader != other._unit_reader || _state.data != other._state.data;
			}

			value_type operator*() { return _state.data; }

			Iterator &operator++() {
				_state = _unit_reader->read();
				return *this;
			}

			Iterator operator++(int) {
				Iterator tmp = *this;
				++(*this);
				return tmp;
			}

			friend Unit;
		};

		Iterator begin() { return Iterator(*this); }
		Iterator end() { return Iterator(*this, { .failed = true }); }
	};

}

#endif
