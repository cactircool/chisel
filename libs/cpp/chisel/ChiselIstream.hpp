#ifndef CHISEL_ISTREAM_HPP
#define CHISEL_ISTREAM_HPP

#include "ChainedStreamBuffer.hpp"
#include <istream>

namespace chisel {

	class ChiselIstream : public std::istream {
	public:
		ChiselIstream(std::istream &stream) : std::istream(new ChainedStreamBuffer({ { .buffer = stream.rdbuf(), .deletable = false } })) {}

		virtual ~ChiselIstream() {
			delete rdbuf();
		}

		void prepend_stream(std::istream &stream) {
			static_cast<ChainedStreamBuffer *>(rdbuf())->prepend_buffer(stream.rdbuf());
			stream.rdbuf(nullptr);
		}

		void append_stream(std::istream &stream) {
			static_cast<ChainedStreamBuffer *>(rdbuf())->append_buffer(stream.rdbuf());
			stream.rdbuf(nullptr);
		}
	};

}

#endif
