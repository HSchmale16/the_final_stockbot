// /**
//  * Helps with the travel calendar to provide better rendering of things
//  */

// // import * as gridjs from 'gridjs';

function renderMemberName(Member) {
    const name = Member.CongressMemberInfo.name.official_full;
    const lastTerm = Member.CongressMemberInfo.terms[Member.CongressMemberInfo.terms.length - 1];

    if (lastTerm.type == "rep") {
        return `${name} (${lastTerm.party}-${lastTerm.state}-${lastTerm.district})`;
    }

    return `${name} (${lastTerm.party}-${lastTerm.state})`;
}

document.addEventListener('DOMContentLoaded', function () {
    function updateFilters() {
        // Generate the url and set the data value
        var url = `/json/travel/calendar/${yearInt}/${monthInt}`;
        if (dateFilter !== null) {
            url = '/json/travel/calendar2?' + new URLSearchParams({ targetDate: dateFilter });
        }

        table.setData(url);

        filtersDIV.innerHTML = '';
        if (dateFilter !== null) {
            const filter = document.createElement('div');
            filter.classList.add('badge');
            filter.classList.add('bg-gray-500');
            filter.innerText = "IncludesDate: " + dateFilter;
            filtersDIV.appendChild(filter);

            var removeBtn = document.createElement('button');
            removeBtn.addEventListener('click', resetDateFilter);
            removeBtn.innerText = '❌';

            filter.appendChild(removeBtn);


            console.log(filtersDIV);
        }

    }

    function resetDateFilter() {
        dateFilter = null;
        updateFilters();
    }

    const travelDays = document.querySelectorAll('.travelDay');
    var filtersDIV = document.getElementById('filtersApplied');
    var dateFilter = null;

    var table = new Tabulator("#wrapper", {
        ajaxURL: `/json/travel/calendar/${yearInt}/${monthInt}`,
        pagination: true,
        // progressiveLoad:"scroll",
        paginationSize:20,
        height: '500px',
        groupBy: "Destination",
        columns: [
            {
                title: "Filer Name",
                field: 'FilerName',
            },
            {
                title: 'Congress Person',
                field: 'Member',
                formatter: (cell, formatterParams) => {
                    var value = cell.getValue();
                    return `<a href="/congress-member/${value.BioGuideId}">${renderMemberName(value)}</a>`;
                 },
            },
            {
                title: 'Sponsor',
                field: 'TravelSponsor',
            },
            {
                title: 'Departure Date',
                field: 'DepartureDate',
                formatter: shittyDateFormat,
            },
            {
                title: 'Return Date',
                field: 'ReturnDate',
                formatter: shittyDateFormat,
            },
            {
                title: 'Destination',
                field: 'Destination'
            },
        ]
    });

    console.log(table);

    travelDays.forEach((day) => {
        day.addEventListener('click', (e) => {
            const day = e.target.innerText;
            const month = e.target.getAttribute('data-month');
            const year = yearInt; // from global scope

            dateFilter = `${year}-${month}-${day}`;
            updateFilters();

            
        });
    });
});

function shittyDateFormat(cell, formatterParams) {
    var value = cell.getValue();
    return new Date(value).toLocaleDateString();
}