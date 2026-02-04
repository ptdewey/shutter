import XCTest
import SwiftTreeSitter
import TreeSitterSnapshot

final class TreeSitterSnapshotTests: XCTestCase {
    func testCanLoadGrammar() throws {
        let parser = Parser()
        let language = Language(language: tree_sitter_snapshot())
        XCTAssertNoThrow(try parser.setLanguage(language),
                         "Error loading Snapshot grammar")
    }
}
