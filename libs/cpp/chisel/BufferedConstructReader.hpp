#ifndef CHISEL_BUFFERED_CONSTRUCT_READER_HPP
#define CHISEL_BUFFERED_CONSTRUCT_READER_HPP

#include "ChiselIstream.hpp"
#include "Unit.hpp"
#include <memory>

namespace chisel {

	class BufferedConstructReader {
		chisel::ChiselIstream *_reader;

	public:
		BufferedConstructReader(chisel::ChiselIstream &reader) : _reader(&reader) {}
		virtual ~BufferedConstructReader() = default;

		chisel::ChiselIstream &reader() { return *_reader; }
		chisel::ChiselIstream &reader() const { return *_reader; }
		void reader(chisel::ChiselIstream &_reader) { this->_reader = &_reader; }

		virtual std::unique_ptr<Unit> next() = 0;

		class Iterator {
            BufferedConstructReader *_parent_reader;
            std::unique_ptr<Unit> _unit_reader;

            // Private constructor for end iterator
            Iterator(BufferedConstructReader *parent_reader, std::unique_ptr<Unit> unit_reader)
                : _parent_reader(parent_reader), _unit_reader(std::move(unit_reader)) {}

        public:
            using iterator_category = std::forward_iterator_tag;
            using difference_type   = std::ptrdiff_t;
            using value_type        = std::unique_ptr<Unit>;
            using pointer           = std::unique_ptr<Unit>*;
            using reference         = std::unique_ptr<Unit>&;

            // Constructor for begin iterator
            Iterator(BufferedConstructReader &parent_reader)
                : _parent_reader(&parent_reader), _unit_reader(parent_reader.next()) {}

            // Copy constructor - note: this moves the unit, making the original invalid
            Iterator(Iterator &&other) noexcept
                : _parent_reader(other._parent_reader), _unit_reader(std::move(other._unit_reader)) {}

            // Move assignment
            Iterator& operator=(Iterator &&other) noexcept {
                if (this != &other) {
                    _parent_reader = other._parent_reader;
                    _unit_reader = std::move(other._unit_reader);
                }
                return *this;
            }

            // Disable copy constructor and copy assignment since we're dealing with unique_ptr
            Iterator(const Iterator &) = delete;
            Iterator& operator=(const Iterator &) = delete;

            reference operator*() { return _unit_reader; }
            pointer operator->() { return &_unit_reader; }

            Iterator &operator++() {
                _unit_reader = _parent_reader->next();
                return *this;
            }

            Iterator operator++(int) {
                Iterator tmp = std::move(*this);
                ++(*this);
                return tmp;
            }

            friend bool operator==(const Iterator &a, const Iterator &b) {
                // Two iterators are equal if they both point to null (end) or both point to valid units
                bool a_is_end = !a._unit_reader;
                bool b_is_end = !b._unit_reader;
                return a_is_end == b_is_end && a._parent_reader == b._parent_reader;
            }

            friend bool operator!=(const Iterator &a, const Iterator &b) {
                return !(a == b);
            }

            friend class BufferedConstructReader;
        };

        Iterator begin() { return Iterator(*this); }
        Iterator end() { return Iterator(this, nullptr); }
	};

}

#endif
