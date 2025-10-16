#include <unordered_map>
#include <cstring>
#include <istream>

namespace chisel {

	class Trie {
		struct Node {
			Token payload;
			bool terminated;
			std::unordered_map<char, Node *> children;
		};

		Node *root;
	public:
		Trie() : root(new Node()) {}
		~Trie() {
			delete root;
		}

		void insert(const char *s, Token &&payload) {
			auto len = strlen(s);
			auto *ptr = root;
			for (auto i = 0; i < len; ++i) {
				auto [it, inserted] = ptr->children.emplace(s[i], nullptr);
				if (inserted)
					it->second = new Node{ .terminated = false };
				ptr = it->second;
			}
			ptr->terminated = true;
			ptr->payload = std::move(payload);
		}

		Token search(const char *s) const {
			auto len = strlen(s);
			auto *ptr = root;
			for (auto i = 0; i < len; ++i) {
				auto it = ptr->children.find(s[i]);
				if (it == ptr->children.end())
					return Token::failed;
				ptr = it->second;
			}
			return ptr->terminated ? ptr->payload : Token::failed;
		}

		// Undoes all changes to the stream for precedence
		Token search(std::istream &r) const {
			auto *ptr = root;
			int count = 0;
			std::pair<Token, int> pair = { Token::failed, 0 };

			while (true) {
				auto c = r.get();
				++count;

				auto it = ptr->children.find(c);
				if (it == ptr->children.end()) {
					if (pair.first) {
						r.seekg(pair.second - count, std::ios::cur);
						return pair.first;
					}
					r.seekg(-count, std::ios::cur);
					break;
				}

				ptr = it->second;
				if (ptr->terminated) pair = { ptr->payload, count };
			}
			return Token::failed;
		}
	};

}
