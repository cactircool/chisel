#ifndef CHISEL_BUFFERED_DATA_HPP
#define CHISEL_BUFFERED_DATA_HPP

namespace chisel {

	struct BufferedData {
		char data = -1;
		bool finished = false;
		bool failed = false;

		bool operator==(const BufferedData &other) const {
			return data == other.data && failed == other.failed && finished == other.finished;
		}

		bool operator!=(const BufferedData &other) const {
			return data != other.data || failed != other.failed || finished != other.finished;
		}
	};

}

#endif
