#include <string>
#include <iostream>

namespace chisel {

	class Result {
		std::string msg;
	public:
		Result() : msg() {}
		Result(const std::string &msg) : msg(msg) {}
		Result(std::string &&msg) : msg(msg) {}
		~Result() = default;

		void panic() const {
			if (msg.empty()) return;
			std::cerr << msg << std::endl;
			exit(1);
		}

		operator bool() const {
			return msg.empty();
		}
	};

}
